package resource_lock

import (
	"github.com/hashicorp/golang-lru"
	"sync"
)

type Lock struct {
	Lock       *lru.Cache
	AccessLock *sync.Mutex
}

func NewResourceLock(size int) *Lock {
	lru, _ := lru.New(size)
	return &Lock{lru, new(sync.Mutex)}
}

func (self *Lock) GetResourceLock(key string) *sync.Mutex {
	self.AccessLock.Lock()
	defer self.AccessLock.Unlock()

	value, ok := self.Lock.Get(key)
	var resourceLock *sync.Mutex
	if ok == false {
		resourceLock = new(sync.Mutex)
		self.Lock.Add(key, resourceLock)
	} else {
		resourceLock = value.(*sync.Mutex)
	}

	return resourceLock
}
