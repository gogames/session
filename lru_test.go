package session

import (
	"testing"
	"time"
)

func Test_LRU(t *testing.T) {
	lru := newLRU()

	lru.put("sessionid1", time.Now().Add(-5*time.Second))
	lru.put("sessionid2", time.Now().Add(-5*time.Second))
	lru.put("sessionid3", time.Now().Add(-5*time.Second))
	lru.put("sessionid4", time.Now().Add(-5*time.Second))
	lru.put("sessionid5", time.Now().Add(-5*time.Second))
	lru.put("sessionid6", time.Now())
	lru.put("sessionid7", time.Now())
	lru.put("sessionid8", time.Now())
	lru.put("sessionid9", time.Now())

	if lru.length() != 9 {
		t.Fatal("length should be 9")
	}

	lru.put("sessionid3", time.Now())
	lru.put("sessionid1", time.Now())
	lru.put("sessionid4", time.Now())

	keys := lru.findExpiredItems(func(val interface{}) bool {
		return time.Since(val.(time.Time)) > 5*time.Second
	})
	lru.remove(keys...)

	if len(keys) != 2 {
		t.Fatal("there should be 2 items removed")
	}

	if lru.length() != 7 {
		t.Fatal("length should be 7")
	}
}
