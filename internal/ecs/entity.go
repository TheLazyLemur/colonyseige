package ecs

import "sync/atomic"

type EntityId uint64

var nextEntityId uint64

type Entity struct {
	Id EntityId
}

func NewEntity() Entity {
	id := atomic.AddUint64(&nextEntityId, 1)
	return Entity{Id: EntityId(id)}
}

func (e Entity) Equal(other Entity) bool {
	return e.Id == other.Id
}
