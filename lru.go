package session

import (
	"bytes"
	"container/list"
	"fmt"
	"sync"
)

// lru used to quick remove expire sessions without scan map
// not a common lru
//
type lru struct {
	l     *list.List
	cache map[interface{}]*list.Element // cache for list remove in O(1)

	lock sync.Mutex
}

// element keep the key for map operation
//
type element struct {
	key interface{}
	val interface{}
}

func newLRU() *lru {
	return &lru{
		l:     list.New(),
		cache: make(map[interface{}]*list.Element),
	}
}

func (l *lru) withLock(f func()) {
	l.lock.Lock()
	defer l.lock.Unlock()
	f()
}

// length
//
func (l *lru) length() int {
	return l.l.Len()
}

// put k-v to back
//
func (l *lru) put(k, v interface{}) {
	element := element{key: k, val: v}
	l.withLock(func() {
		// exist, move to back
		if e, ok := l.cache[k]; ok {
			e.Value = element
			l.l.MoveToBack(e)
			return
		}
		// not exist, push to back
		l.cache[k] = l.l.PushBack(element)
	})
}

// remove items
//
func (l *lru) remove(ks ...interface{}) {
	l.withLock(func() {
		for _, k := range ks {
			if e, ok := l.cache[k]; ok {
				l.l.Remove(e)
				delete(l.cache, k)
			}
		}
	})
}

// iterate lru, from front to back, find expired items
//
func (l *lru) findExpiredItems(isExpired func(value interface{}) bool) []interface{} {
	expiredItems := make([]interface{}, 0)

	l.withLock(func() {
		for e := l.l.Front(); e != nil; e = e.Next() {
			// if expired, put it in slice
			if isExpired(e.Value.(element).val) {
				expiredItems = append(expiredItems, e.Value.(element).key)
			}
		}
	})

	return expiredItems
}

// print lru
//
func (l *lru) String() string {
	s := bytes.NewBufferString("\n")
	for e := l.l.Front(); e != nil; e = e.Next() {
		s.WriteString(fmt.Sprintf("key: %v\tvalue: %v\n", e.Value.(element).key, e.Value.(element).val))
	}
	return s.String()
}
