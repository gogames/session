package session

import (
	"testing"
	"time"
)

type specialType struct {
	A *specialType
}

func Test_Session(t *testing.T) {
	test(t, fileSession())
	test(t, memorySession())
}

func test(t *testing.T, s Session) {
	sid, key, val := s.GenerateID(), "helloKey", specialType{}

	if err := s.Set(sid, key, val); err != nil {
		t.Fatal(err)
	}

	if s.Get(sid, key).(specialType) != val {
		t.Fatalf("the value is not %v", val)
	}

	time.Sleep(1200 * time.Millisecond)

	if value := s.Get(sid, key); value != nil {
		t.Fatalf("the value should be nil interface{} but get %v", value)
	}

	nsid := s.GenerateID()
	err := s.Set(nsid, "ha", "lo")
	if err != nil {
		t.Fatal(err)
	}
	if s.Get(nsid, "ha").(string) != "lo" {
		t.Fatal("should be lo")
	}

	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}

	if err := s.Set(nsid, "what", "the"); err == nil {
		t.Fatal("error should be nil")
	}
}

func fileSession() Session {
	return NewSession(NewFileStore(nil, "dir", "/", 1*time.Second), 50)
}

func memorySession() Session {
	return NewSession(NewMemoryStore(nil, 1*time.Second), 50)
}
