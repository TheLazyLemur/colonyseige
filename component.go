package ecs

import (
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

type entityComponents struct {
	values  []Component
	present []bool
	count   int
}

func newEntityComponents() *entityComponents {
	return &entityComponents{
		values:  make([]Component, 0),
		present: make([]bool, 0),
	}
}

func componentIndex(id ComponentId) int {
	if id == 0 {
		return -1
	}
	return int(id) - 1
}

func componentIdFromIndex(idx int) ComponentId {
	return ComponentId(idx + 1)
}

func (ec *entityComponents) ensureCapacity(idx int) {
	if idx < len(ec.values) {
		return
	}

	newLen := len(ec.values)
	if newLen == 0 {
		newLen = 1
	}
	for newLen <= idx {
		newLen *= 2
	}

	newValues := make([]Component, newLen)
	copy(newValues, ec.values)
	newPresent := make([]bool, newLen)
	copy(newPresent, ec.present)

	ec.values = newValues
	ec.present = newPresent
}

func (ec *entityComponents) set(id ComponentId, component Component) bool {
	idx := componentIndex(id)
	if idx < 0 {
		return false
	}

	ec.ensureCapacity(idx)
	wasPresent := ec.present[idx]
	ec.values[idx] = component
	ec.present[idx] = true
	if !wasPresent {
		ec.count++
	}

	return !wasPresent
}

func (ec *entityComponents) get(id ComponentId) (Component, bool) {
	idx := componentIndex(id)
	if idx < 0 || idx >= len(ec.present) || !ec.present[idx] {
		return nil, false
	}

	return ec.values[idx], true
}

func (ec *entityComponents) has(id ComponentId) bool {
	idx := componentIndex(id)
	return idx >= 0 && idx < len(ec.present) && ec.present[idx]
}

func (ec *entityComponents) remove(id ComponentId) bool {
	idx := componentIndex(id)
	if idx < 0 || idx >= len(ec.present) || !ec.present[idx] {
		return false
	}

	ec.present[idx] = false
	ec.values[idx] = nil
	ec.count--
	return true
}

func (ec *entityComponents) toMap() map[ComponentId]Component {
	result := make(map[ComponentId]Component, ec.count)
	for idx, present := range ec.present {
		if !present {
			continue
		}
		result[componentIdFromIndex(idx)] = ec.values[idx]
	}
	return result
}

type componentEntry struct {
	id        ComponentId
	component Component
}

func (ec *entityComponents) entries() []componentEntry {
	if ec.count == 0 {
		return nil
	}
	entries := make([]componentEntry, 0, ec.count)
	for idx, present := range ec.present {
		if !present {
			continue
		}
		entries = append(entries, componentEntry{
			id:        componentIdFromIndex(idx),
			component: ec.values[idx],
		})
	}
	return entries
}

type ComponentStore struct {
	components        map[EntityId]*entityComponents
	componentEntities map[ComponentId]map[EntityId]struct{}
	mutex             sync.RWMutex
}

func NewComponentStore() *ComponentStore {
	return &ComponentStore{
		components:        make(map[EntityId]*entityComponents),
		componentEntities: make(map[ComponentId]map[EntityId]struct{}),
	}
}

func (cs *ComponentStore) AddComponent(entityId EntityId, component Component) {
	componentId := GetComponentId(component)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	ec := cs.components[entityId]
	if ec == nil {
		ec = newEntityComponents()
		cs.components[entityId] = ec
	}

	ec.set(componentId, component)
	if cs.componentEntities[componentId] == nil {
		cs.componentEntities[componentId] = make(map[EntityId]struct{})
	}
	cs.componentEntities[componentId][entityId] = struct{}{}
}

func (cs *ComponentStore) GetComponent(entityId EntityId, componentType reflect.Type) (Component, bool) {
	return cs.GetComponentById(entityId, getComponentIdByType(componentType))
}

func (cs *ComponentStore) RemoveComponent(entityId EntityId, componentType reflect.Type) {
	cs.RemoveComponentById(entityId, getComponentIdByType(componentType))
}

func (cs *ComponentStore) HasComponent(entityId EntityId, componentType reflect.Type) bool {
	return cs.HasComponentById(entityId, getComponentIdByType(componentType))
}

func (cs *ComponentStore) GetEntityComponents(entityId EntityId) map[ComponentId]Component {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	ec := cs.components[entityId]
	if ec == nil {
		return make(map[ComponentId]Component)
	}

	return ec.toMap()
}

func (cs *ComponentStore) GetComponentById(entityId EntityId, componentId ComponentId) (Component, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	ec := cs.components[entityId]
	if ec == nil {
		return nil, false
	}

	return ec.get(componentId)
}

func (cs *ComponentStore) HasComponentById(entityId EntityId, componentId ComponentId) bool {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	ec := cs.components[entityId]
	if ec == nil {
		return false
	}

	return ec.has(componentId)
}

func (cs *ComponentStore) RemoveComponentById(entityId EntityId, componentId ComponentId) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	ec := cs.components[entityId]
	if ec == nil {
		return
	}

	if !ec.remove(componentId) {
		return
	}

	if bucket, ok := cs.componentEntities[componentId]; ok {
		delete(bucket, entityId)
		if len(bucket) == 0 {
			delete(cs.componentEntities, componentId)
		}
	}

	if ec.count == 0 {
		delete(cs.components, entityId)
	}
}

func (cs *ComponentStore) hasAll(entityId EntityId, ids []ComponentId) bool {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return cs.hasAllUnlocked(entityId, ids)
}

func (cs *ComponentStore) hasAllUnlocked(entityId EntityId, ids []ComponentId) bool {
	ec := cs.components[entityId]
	if ec == nil {
		return false
	}

	for _, id := range ids {
		if !ec.has(id) {
			return false
		}
	}

	return true
}

func (cs *ComponentStore) IterateEntityComponents(entityId EntityId, fn func(ComponentId, Component) bool) {
	cs.mutex.RLock()
	entries := []componentEntry{}
	if ec := cs.components[entityId]; ec != nil {
		entries = ec.entries()
	}
	cs.mutex.RUnlock()

	for _, entry := range entries {
		if !fn(entry.id, entry.component) {
			break
		}
	}
}

func (cs *ComponentStore) RemoveAllComponents(entityId EntityId) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	ec := cs.components[entityId]
	if ec == nil {
		return
	}

	for idx, present := range ec.present {
		if !present {
			continue
		}
		componentId := componentIdFromIndex(idx)
		if bucket, ok := cs.componentEntities[componentId]; ok {
			delete(bucket, entityId)
			if len(bucket) == 0 {
				delete(cs.componentEntities, componentId)
			}
		}
	}

	delete(cs.components, entityId)
}

func (cs *ComponentStore) EntitiesWithAll(ids []ComponentId) []EntityId {
	if len(ids) == 0 {
		return nil
	}

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return cs.entitiesWithAllLocked(ids)
}

func (cs *ComponentStore) entitiesWithAllLocked(ids []ComponentId) []EntityId {
	var baseID ComponentId
	var baseEntities map[EntityId]struct{}

	for _, id := range ids {
		entities, ok := cs.componentEntities[id]
		if !ok || len(entities) == 0 {
			return []EntityId{}
		}

		if baseEntities == nil || len(entities) < len(baseEntities) {
			baseID = id
			baseEntities = entities
		}
	}

	result := make([]EntityId, 0, len(baseEntities))

	for entityId := range baseEntities {
		ec := cs.components[entityId]
		if ec == nil {
			continue
		}

		match := true
		for _, id := range ids {
			if id == baseID {
				continue
			}
			if !ec.has(id) {
				match = false
				break
			}
		}

		if match {
			result = append(result, entityId)
		}
	}

	return result
}
