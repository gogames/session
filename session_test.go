package session

import (
	"encoding/json"
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

func test(t *testing.T, s *Session) {
	sid, key, val := "", "helloKey", specialType{}

	sid, err := s.Set(sid, key, val)
	if err != nil {
		t.Fatal(err)
	}

	if s.Get(sid, key).(specialType) != val {
		t.Fatalf("the value is not %v", val)
	}

	time.Sleep(5 * time.Millisecond)

	if value := s.Get(sid, key); value != nil {
		t.Fatalf("the value should be nil interface{} but get %v", value)
	}

	nsid, err := s.Set("", "ha", "lo")
	if err != nil {
		t.Fatal(err)
	}
	if s.Get(nsid, "ha").(string) != "lo" {
		t.Fatal("should be lo")
	}

	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}

	s.Close()

	if nsid, _ = s.Set(nsid, "what", "the"); nsid != "" {
		t.Fatal("still can set after session is close")
	}
}

func fileSession() *Session {
	config, _ := json.Marshal(map[string]interface{}{
		"path":      "dir",
		"separator": "/",
	})

	return NewSession(STORE_FILE, time.Millisecond, time.Millisecond, string(config))
}

func memorySession() *Session {
	return NewSession(STORE_MEMORY, time.Millisecond, time.Millisecond, "")
}
