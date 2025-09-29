package ecs

import "reflect"

type System interface {
	Update(world *World, deltaTime float64)
	RequiredComponents() []reflect.Type
}

type Filter struct {
	requiredComponents []ComponentId
}

func NewFilter(componentTypes ...reflect.Type) *Filter {
	requiredComponents := make([]ComponentId, len(componentTypes))
	for i, componentType := range componentTypes {
		requiredComponents[i] = getComponentIdByType(componentType)
	}

	return &Filter{
		requiredComponents: requiredComponents,
	}
}

func (f *Filter) Matches(entityId EntityId, componentStore *ComponentStore) bool {
	return componentStore.hasAll(entityId, f.requiredComponents)
}

func (f *Filter) FilterEntities(entities []Entity, componentStore *ComponentStore) []Entity {
	filtered := make([]Entity, 0, len(entities))

	for _, entity := range entities {
		if f.Matches(entity.Id, componentStore) {
			filtered = append(filtered, entity)
		}
	}

	return filtered
}
