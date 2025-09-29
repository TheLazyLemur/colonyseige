package main

import (
	"fmt"
	"reflect"

	"ecs"
)

var (
	positionType = reflect.TypeOf(Position{})
	velocityType = reflect.TypeOf(Velocity{})
	healthType   = reflect.TypeOf(Health{})
)

type MovementSystem struct {
	positionId ecs.ComponentId
	velocityId ecs.ComponentId
}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{
		positionId: ecs.GetComponentId(Position{}),
		velocityId: ecs.GetComponentId(Velocity{}),
	}
}

func (ms *MovementSystem) Update(world *ecs.World, deltaTime float64) {
	entities := world.QueryEntities(positionType, velocityType)

	for _, entity := range entities {
		posComp, _ := world.GetComponentById(entity, ms.positionId)
		velComp, _ := world.GetComponentById(entity, ms.velocityId)

		pos := posComp.(Position)
		vel := velComp.(Velocity)

		pos.X += vel.X * deltaTime
		pos.Y += vel.Y * deltaTime

		world.AddComponent(entity, pos)
	}
}

func (ms *MovementSystem) RequiredComponents() []reflect.Type {
	return []reflect.Type{positionType, velocityType}
}

type HealthSystem struct {
	healthId ecs.ComponentId
}

func NewHealthSystem() *HealthSystem {
	return &HealthSystem{
		healthId: ecs.GetComponentId(Health{}),
	}
}

func (hs *HealthSystem) Update(world *ecs.World, deltaTime float64) {
	entities := world.QueryEntities(healthType)

	for _, entity := range entities {
		healthComp, _ := world.GetComponentById(entity, hs.healthId)
		health := healthComp.(Health)

		if health.Current <= 0 {
			fmt.Printf("Entity %d has died!\n", entity.Id)
			world.DestroyEntity(entity)
		}
	}
}

func (hs *HealthSystem) RequiredComponents() []reflect.Type {
	return []reflect.Type{healthType}
}
