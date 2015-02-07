package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	_SID_LENGTH   = 1 << 4  // session id length
	_EMPTY_STRING = ""      // used to check empty string
	_BUFFER       = 1 << 16 // at most process _BUFFER job in the mean time
)

var defaultGCFrequency = time.Second

// a session store should provide these functions
//
type SessionStore interface {
	Set(string, interface{}) error
	Get(string) interface{}
	Delete(string) error
	Update() error
	LastUpdate() time.Time
	SessionId() string
	Expire() error
	Init() map[string]time.Time
}

// entries
//
var sessions = make(map[string]func(sid string, conf string) SessionStore)

// register a session store
//
func Register(sessionStoreName string, f func(string, string) SessionStore) {
	if _, ok := sessions[sessionStoreName]; ok {
		panic(fmt.Sprintf("can not register %v twice", sessionStoreName))
	}
	sessions[sessionStoreName] = f
}

// thread safe session
//
type Session struct {
	sessions map[string]SessionStore

	storeProvider func(string, string) SessionStore

	// used for quick gc, not scan all session any more
	lru *lru

	// gc
	gcFrequency     time.Duration
	sessionLifeTime time.Duration

	closeSignal chan struct{}
	closeChan   chan struct{}
	isClosed    bool

	conf string

	rwl sync.RWMutex
}

func NewSession(sessionStoreName string, gcFrequency, sessionLifeTime time.Duration, conf string) *Session {
	s := &Session{
		sessionLifeTime: sessionLifeTime,
		gcFrequency:     gcFrequency,
		closeSignal:     make(chan struct{}),
		closeChan:       make(chan struct{}, _BUFFER),
		conf:            conf,
		sessions:        make(map[string]SessionStore),
		lru:             newLRU(),
	}

	// choose a store
	if sp, ok := sessions[sessionStoreName]; !ok {
		panic(fmt.Sprintf("session store %v is not registered", sessionStoreName))
	} else {
		s.storeProvider = sp
		for sessionId, lastUpdate := range sp(_EMPTY_STRING, conf).Init() {
			s.lru.put(sessionId, lastUpdate)
			s.sessions[sessionId] = s.storeProvider(sessionId, conf)
		}
	}

	// start gc
	go s.gc()

	return s
}

// waits for all operations to finish
// usually used with graceful exit, ensure data integrity
//
func (s *Session) Close() {
	s.withWriteLock(func() { s.isClosed = true })
	for len(s.closeChan) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	s.closeSignal <- struct{}{}
}

// remove all sessions
// if session is closed, do nothing
//
func (s *Session) Flush() (err error) {
	s.do(func() {
		s.withWriteLock(func() {
			for sid, ss := range s.sessions {
				if err = ss.Expire(); err != nil {
					return
				}
				delete(s.sessions, sid)
				s.lru.remove(sid)
			}
		})
	})
	return
}

// get a key from session store
// if session is closed, do nothing
//
func (s *Session) Get(sid string, key string) (val interface{}) {
	if sid == _EMPTY_STRING {
		return
	}
	s.withReadLock(func() {
		if sp, ok := s.sessions[sid]; ok {
			val = sp.Get(key)
		}
	})
	return
}

// set key-value pair
// if the sid is empty string, new generated sessionId will be returned
// if session is closed, do nothing
//
func (s *Session) Set(sid string, key string, val interface{}) (sessionId string, err error) {
	s.do(func() {
		var cont bool
		if sid != _EMPTY_STRING {
			s.withReadLock(func() {
				sp, ok := s.sessions[sid]
				if ok {
					cont = ok
					err = sp.Set(key, val)
				}
			})
		}
		if !cont {
			s.withWriteLock(func() {
				sid = s.newSId()
				s.sessions[sid] = s.storeProvider(sid, s.conf)
				err = s.sessions[sid].Set(key, val)
			})
		}
		sessionId = sid
		s.lru.put(sessionId, time.Now())
	})
	return
}

// delete a key-value pair
// if session is closed, do nothing
//
func (s *Session) Delete(sid string, key string) (err error) {
	if sid == _EMPTY_STRING {
		return
	}
	s.do(func() {
		s.withReadLock(func() {
			if sp, ok := s.sessions[sid]; ok {
				err = sp.Delete(key)
			}
		})
	})
	return
}

// update session life time according to session id
// if session is closed, do nothing
//
func (s *Session) Update(sid string) (err error) {
	if sid == _EMPTY_STRING {
		return
	}
	s.do(func() {
		s.withReadLock(func() {
			sp, ok := s.sessions[sid]
			if ok {
				err = sp.Update()
			}
		})
	})
	return
}

// expire the session according to session id
// if session is closed, do nothing
//
func (s *Session) Expire(sid string) (err error) {
	if sid == _EMPTY_STRING {
		return
	}
	s.do(func() {
		s.withWriteLock(func() {
			if ss, ok := s.sessions[sid]; ok {
				err = ss.Expire()
			}
			delete(s.sessions, sid)
			s.lru.remove(sid)
		})
	})
	return
}

func (s *Session) gc() {
	for {
		select {
		case t := <-time.Tick(s.gcFrequency):
			s.do(func() {
				s.withWriteLock(func() {
					sids := s.lru.findExpiredItems(func(value interface{}) bool {
						return t.Sub(value.(time.Time)) > s.sessionLifeTime
					})
					s.lru.remove(sids...)

					for _, sidInterface := range sids {
						sid := sidInterface.(string)
						s.sessions[sid].Expire()
						delete(s.sessions, sid)
					}
				})
			})
		case <-s.closeSignal:
			return
		}
	}
}

func (s *Session) withReadLock(f func()) {
	s.rwl.RLock()
	defer s.rwl.RUnlock()
	f()
}

func (s *Session) withWriteLock(f func()) {
	s.rwl.Lock()
	defer s.rwl.Unlock()
	f()
}

func (s *Session) do(f func()) {
	if s.closed() {
		return
	}
	s.closeChan <- struct{}{}
	defer func() { <-s.closeChan }()
	f()
}

// check if closed
func (s *Session) closed() bool {
	s.rwl.RLock()
	defer s.rwl.RUnlock()
	return s.isClosed
}

func (s *Session) newSId() string {
	for {
		if sid := s.randSId(); sid != "" {
			if _, ok := s.sessions[sid]; ok {
				continue
			}
			return sid
		}
	}
}

func (s *Session) randSId() string {
	for {
		b := make([]byte, _SID_LENGTH)
		n, err := rand.Read(b)
		if err != nil && n != _SID_LENGTH {
			log.Printf("can not rand.Read %v", err)
			return ""
		}
		return hex.EncodeToString(b)
	}
}
