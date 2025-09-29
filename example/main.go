package main

import (
	"fmt"
	"time"

	"ecs"
)

func main() {
	world := ecs.NewWorld()

	movementSystem := NewMovementSystem()
	healthSystem := NewHealthSystem()

	world.AddSystem(movementSystem)
	world.AddSystem(healthSystem)

	player := world.CreateEntity()
	world.AddComponent(player, Position{X: 0, Y: 0})
	world.AddComponent(player, Velocity{X: 10, Y: 5})
	world.AddComponent(player, Health{Current: 100, Max: 100})

	enemy := world.CreateEntity()
	world.AddComponent(enemy, Position{X: 50, Y: 50})
	world.AddComponent(enemy, Velocity{X: -5, Y: -2})
	world.AddComponent(enemy, Health{Current: 0, Max: 50})

	fmt.Println("Starting game loop...")

	positionId := ecs.GetComponentId(Position{})

	for i := range 20 {
		deltaTime := 0.016

		fmt.Printf("\n--- Frame %d ---\n", i+1)
		if i == 10 {
			player2 := world.CreateEntity()
			world.AddComponent(player2, Position{X: 0, Y: 0})
			world.AddComponent(player2, Velocity{X: 0, Y: 20})
			world.AddComponent(player2, Health{Current: 100, Max: 100})
		}

		entities := world.GetEntities()
		for _, entity := range entities {
			if pos, exists := world.GetComponentById(entity, positionId); exists {
				position := pos.(Position)
				fmt.Printf("Entity %d position: (%.2f, %.2f)\n", entity.Id, position.X, position.Y)
			}
		}

		world.Update(deltaTime)

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\nGame loop finished!")
}
