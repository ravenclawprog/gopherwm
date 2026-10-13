package manager

import (
	"errors"
	"gopherwm/idgen"
	"sync"
)

type ObjectManager[T any] struct {
	mutex       sync.Mutex
	objects     map[uint64]T
	idGenerator *idgen.IDGenerator
}

func NewObjectManager[T any]() *ObjectManager[T] {
	return &ObjectManager[T]{
		objects:     make(map[uint64]T),
		idGenerator: idgen.NewIDGenerator(),
	}
}

func (om *ObjectManager[T]) GetUnknownKey() uint64 {
	unknwn := uint64(0)
	return unknwn
}

func (om *ObjectManager[T]) AddObject(value T) uint64 {
	defer om.mutex.Unlock()
	om.mutex.Lock()
	newId := om.idGenerator.Generate()
	if _, ok := om.objects[newId]; ok {
		return om.GetUnknownKey()
	}
	om.objects[newId] = value
	return newId
}

func (om *ObjectManager[T]) GetObject(index uint64) (T, error) {
	defer om.mutex.Unlock()
	om.mutex.Lock()
	if _, ok := om.objects[index]; !ok {
		var val T
		return val, errors.New("Cannot find object")
	}
	return om.objects[index], nil
}
