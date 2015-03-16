// memory store
package session

import (
	"fmt"
	"sync"
	"time"
)

var _ SessionStore = new(memory)

type memoryElement struct {
	data       map[string]interface{}
	lastUpdate time.Time
}

type memory struct {
	data       map[string]*memoryElement
	generateID func() string
	rwl        sync.RWMutex
}

func NewMemoryStore(IDGenerator func() string) *memory {
	if IDGenerator == nil {
		IDGenerator = DefaultGenerator
	}
	return &memory{make(map[string]*memoryElement), IDGenerator, sync.RWMutex{}}
}

func (m *memory) withReadLock(f func()) {
	m.rwl.RLock()
	defer m.rwl.RUnlock()
	f()
}

func (m *memory) withWriteLock(f func()) {
	m.rwl.Lock()
	defer m.rwl.Unlock()
	f()
}

func (m *memory) Expire(ID string) error {
	m.withWriteLock(func() {
		delete(m.data, ID)
	})
	return nil
}

func (m *memory) Update(ID string) error {
	m.withWriteLock(func() {
		if d, ok := m.data[ID]; ok {
			d.lastUpdate = time.Now()
		}
	})
	return nil
}

var errSessionFlushed = fmt.Errorf("session flushed")

func (m *memory) Set(ID string, key string, val interface{}) (err error) {
	m.withWriteLock(func() {
		if m.data == nil {
			err = errSessionFlushed
			return
		}
		if d, ok := m.data[ID]; ok {
			d.data[key] = val
		}
	})
	return
}

func (m *memory) Get(ID string, key string) (val interface{}) {
	m.withReadLock(func() {
		if d, ok := m.data[ID]; ok {
			val = d.data[key]
		}
	})
	return
}

func (m *memory) Delete(ID string, key string) error {
	m.withWriteLock(func() {
		if d, ok := m.data[ID]; ok {
			delete(d.data, key)
		}
	})
	return nil
}

func (m *memory) Flush() error {
	m.withWriteLock(func() {
		m.data = nil
	})
	return nil
}

func (m *memory) GC(lifeTime time.Duration, t time.Time) {
	m.withWriteLock(func() {
		for ID, d := range m.data {
			if d.lastUpdate.Add(lifeTime).Before(t) {
				delete(m.data, ID)
			}
		}
	})
}

func (m *memory) GenerateID() (id string) {
	m.withWriteLock(func() {
		for {
			id = m.generateID()
			if _, ok := m.data[id]; ok {
				continue
			}
			m.data[id] = &memoryElement{make(map[string]interface{}), time.Now()}
			break
		}
	})
	return
}
