package ecs

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type World struct {
	entities       []Entity
	entityIndex    map[EntityId]int
	componentStore *ComponentStore
	systems        []System
	systemsCache   atomic.Pointer[[]System]
	mutex          sync.RWMutex
}

func NewWorld() *World {
	return &World{
		entities:       make([]Entity, 0),
		entityIndex:    make(map[EntityId]int),
		componentStore: NewComponentStore(),
		systems:        make([]System, 0),
	}
}

func (w *World) CreateEntity() Entity {
	entity := NewEntity()

	w.mutex.Lock()
	w.entityIndex[entity.Id] = len(w.entities)
	w.entities = append(w.entities, entity)
	w.mutex.Unlock()

	return entity
}

func (w *World) DestroyEntity(entity Entity) {
	var entityToRemove EntityId

	w.mutex.Lock()
	index, exists := w.entityIndex[entity.Id]
	if !exists {
		w.mutex.Unlock()
		return
	}

	lastIndex := len(w.entities) - 1
	if index != lastIndex {
		lastEntity := w.entities[lastIndex]
		w.entities[index] = lastEntity
		w.entityIndex[lastEntity.Id] = index
	}

	w.entities = w.entities[:lastIndex]
	delete(w.entityIndex, entity.Id)
	entityToRemove = entity.Id
	w.mutex.Unlock()

	w.componentStore.RemoveAllComponents(entityToRemove)
}

func (w *World) AddComponent(entity Entity, component Component) {
	w.componentStore.AddComponent(entity.Id, component)
}

func (w *World) GetComponent(entity Entity, componentType reflect.Type) (Component, bool) {
	return w.componentStore.GetComponent(entity.Id, componentType)
}

func (w *World) RemoveComponent(entity Entity, componentType reflect.Type) {
	w.componentStore.RemoveComponent(entity.Id, componentType)
}

func (w *World) HasComponent(entity Entity, componentType reflect.Type) bool {
	return w.componentStore.HasComponent(entity.Id, componentType)
}

func (w *World) GetComponentById(entity Entity, componentId ComponentId) (Component, bool) {
	return w.componentStore.GetComponentById(entity.Id, componentId)
}

func (w *World) HasComponentById(entity Entity, componentId ComponentId) bool {
	return w.componentStore.HasComponentById(entity.Id, componentId)
}

func (w *World) RemoveComponentById(entity Entity, componentId ComponentId) {
	w.componentStore.RemoveComponentById(entity.Id, componentId)
}

func (w *World) IterateEntityComponents(entity Entity, fn func(ComponentId, Component) bool) {
	w.componentStore.IterateEntityComponents(entity.Id, fn)
}

func (w *World) AddSystem(system System) {
	w.mutex.Lock()
	w.systems = append(w.systems, system)
	w.refreshSystemsCacheLocked()
	w.mutex.Unlock()
}

func (w *World) refreshSystemsCacheLocked() {
	snapshot := append([]System(nil), w.systems...)
	ptr := new([]System)
	*ptr = snapshot
	w.systemsCache.Store(ptr)
}

func (w *World) Update(deltaTime float64) {
	systemsPtr := w.systemsCache.Load()
	if systemsPtr == nil {
		w.mutex.RLock()
		snapshot := append([]System(nil), w.systems...)
		w.mutex.RUnlock()
		ptr := new([]System)
		*ptr = snapshot
		if !w.systemsCache.CompareAndSwap(nil, ptr) {
			systemsPtr = w.systemsCache.Load()
		} else {
			systemsPtr = ptr
		}
	}
	if systemsPtr == nil {
		return
	}

	for _, system := range *systemsPtr {
		system.Update(w, deltaTime)
	}
}

func (w *World) QueryEntities(componentTypes ...reflect.Type) []Entity {
	return w.FilterEntities(NewFilter(componentTypes...))
}

func (w *World) FilterEntities(filter *Filter) []Entity {
	if len(filter.requiredComponents) == 0 {
		w.mutex.RLock()
		entities := append([]Entity(nil), w.entities...)
		w.mutex.RUnlock()
		return entities
	}

	ids := w.componentStore.EntitiesWithAll(filter.requiredComponents)
	if len(ids) == 0 {
		return []Entity{}
	}

	w.mutex.RLock()
	defer w.mutex.RUnlock()

	filtered := make([]Entity, 0, len(ids))
	for _, id := range ids {
		if index, exists := w.entityIndex[id]; exists {
			filtered = append(filtered, w.entities[index])
		}
	}

	return filtered
}

func (w *World) GetEntities() []Entity {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	entities := make([]Entity, len(w.entities))
	copy(entities, w.entities)
	return entities
}
