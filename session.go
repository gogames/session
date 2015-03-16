package session

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"
)

type SessionStore interface {
	GenerateID() string
	Set(ID string, key string, val interface{}) error
	Get(ID string, key string) interface{}
	Delete(ID string, key string) error
	Update(ID string) error
	Expire(ID string) error
	Flush() error
	GC(lifeTime time.Duration, timeNow time.Time)
}

type Session struct {
	SessionStore
	lifeTime                 time.Duration
	gcFrequencyInMilliSecond int64
}

func NewSession(store SessionStore, sessionLifeTime time.Duration, gcFrequencyInMilliSecond int64) Session {
	s := Session{store, sessionLifeTime, gcFrequencyInMilliSecond}
	go s.gc()
	return s
}

func (s Session) gc() {
	if s.gcFrequencyInMilliSecond <= 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(s.gcFrequencyInMilliSecond) * time.Millisecond)
	for {
		select {
		case t := <-ticker.C:
			s.GC(s.lifeTime, t)
		}
	}
}

// DefaultGenerator generate 16 bytes session id
var DefaultGenerator = func() string {
	const length = 16
	b := make([]byte, length)
	n, err := rand.Read(b)
	if err != nil && n != length {
		log.Fatalf("can not rand.Read %v", err)
	}
	return hex.EncodeToString(b)
}
