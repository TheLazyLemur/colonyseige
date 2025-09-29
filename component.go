package ecs

import (
	"maps"
	"reflect"
	"sync"
)

type ComponentId uint32

var (
	componentIdCounter uint32
	componentIdMutex   sync.RWMutex
	componentTypeToId  = make(map[reflect.Type]ComponentId)
)

type Component interface{}

func GetComponentId(component Component) ComponentId {
	return getComponentIdByType(reflect.TypeOf(component))
}

func getComponentIdByType(componentType reflect.Type) ComponentId {
	componentIdMutex.RLock()
	id, exists := componentTypeToId[componentType]
	componentIdMutex.RUnlock()

	if exists {
		return id
	}

	componentIdMutex.Lock()
	defer componentIdMutex.Unlock()

	if id, exists := componentTypeToId[componentType]; exists {
		return id
	}

	componentIdCounter++
	id = ComponentId(componentIdCounter)
	componentTypeToId[componentType] = id

	return id
}

type ComponentStore struct {
	components map[EntityId]map[ComponentId]Component
	mutex      sync.RWMutex
}

func NewComponentStore() *ComponentStore {
	return &ComponentStore{
		components: make(map[EntityId]map[ComponentId]Component),
	}
}

func (cs *ComponentStore) AddComponent(entityId EntityId, component Component) {
	componentId := GetComponentId(component)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if cs.components[entityId] == nil {
		cs.components[entityId] = make(map[ComponentId]Component, 4)
	}

	cs.components[entityId][componentId] = component
}

func (cs *ComponentStore) GetComponent(entityId EntityId, componentType reflect.Type) (Component, bool) {
	componentId := getComponentIdByType(componentType)

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	if entityComponents, exists := cs.components[entityId]; exists {
		if component, exists := entityComponents[componentId]; exists {
			return component, true
		}
	}

	return nil, false
}

func (cs *ComponentStore) RemoveComponent(entityId EntityId, componentType reflect.Type) {
	componentId := getComponentIdByType(componentType)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if entityComponents, exists := cs.components[entityId]; exists {
		delete(entityComponents, componentId)

		if len(entityComponents) == 0 {
			delete(cs.components, entityId)
		}
	}
}

func (cs *ComponentStore) HasComponent(entityId EntityId, componentType reflect.Type) bool {
	_, exists := cs.GetComponent(entityId, componentType)
	return exists
}

func (cs *ComponentStore) GetEntityComponents(entityId EntityId) map[ComponentId]Component {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	if components, exists := cs.components[entityId]; exists {
		result := make(map[ComponentId]Component, len(components))
		maps.Copy(result, components)
		return result
	}

	return make(map[ComponentId]Component)
}

func (cs *ComponentStore) GetComponentById(entityId EntityId, componentId ComponentId) (Component, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	if entityComponents, exists := cs.components[entityId]; exists {
		if component, exists := entityComponents[componentId]; exists {
			return component, true
		}
	}

	return nil, false
}

func (cs *ComponentStore) HasComponentById(entityId EntityId, componentId ComponentId) bool {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	if entityComponents, exists := cs.components[entityId]; exists {
		_, exists := entityComponents[componentId]
		return exists
	}

	return false
}

func (cs *ComponentStore) RemoveComponentById(entityId EntityId, componentId ComponentId) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if entityComponents, exists := cs.components[entityId]; exists {
		delete(entityComponents, componentId)

		if len(entityComponents) == 0 {
			delete(cs.components, entityId)
		}
	}
}

func (cs *ComponentStore) hasAll(entityId EntityId, ids []ComponentId) bool {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return cs.hasAllUnlocked(entityId, ids)
}

func (cs *ComponentStore) IterateEntityComponents(entityId EntityId, fn func(ComponentId, Component) bool) {
	cs.mutex.RLock()
	components, exists := cs.components[entityId]
	if !exists {
		cs.mutex.RUnlock()
		return
	}

	componentsCopy := make([]struct {
		id        ComponentId
		component Component
	}, 0, len(components))

	for id, component := range components {
		componentsCopy = append(componentsCopy, struct {
			id        ComponentId
			component Component
		}{id, component})
	}
	cs.mutex.RUnlock()

	for _, item := range componentsCopy {
		if !fn(item.id, item.component) {
			break
		}
	}
}

func (cs *ComponentStore) RemoveAllComponents(entityId EntityId) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.components, entityId)
}

func (cs *ComponentStore) hasAllUnlocked(entityId EntityId, ids []ComponentId) bool {
	entityComponents, exists := cs.components[entityId]
	if !exists {
		return false
	}

	for _, id := range ids {
		if _, exists := entityComponents[id]; !exists {
			return false
		}
	}

	return true
}
