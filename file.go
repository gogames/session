// file store
package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const STORE_FILE = "file"

type file struct {
	path          string
	pathSeparator string
	sid           string
}

func newFileStore(sid string, conf string) SessionStore {
	f := &file{sid: sid}
	f.parseConf(conf)
	if sid != _EMPTY_STRING {
		if err := os.Mkdir(f.getDir(), os.ModePerm); err != nil {
			panic(err)
		}
	}
	return f
}

func (f *file) SessionId() string {
	return f.sid
}

func (f *file) Init() {
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		if err = os.Mkdir(f.path, os.ModePerm); err != nil {
			panic(fmt.Errorf("can not mkdir: %v", err))
		}
	}
}

func (f *file) Iterate(fc func(key string, val interface{})) {
	filepath.Walk(f.getDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// get key and value
		b, err := ioutil.ReadFile(info.Name())
		if err != nil {
			return err
		}
		fc(info.Name(), unmarshal(b))
		return nil
	})
}

// set value
func (f *file) Set(key string, val interface{}) error {
	return ioutil.WriteFile(f.getFileName(key), marshal(val), os.ModePerm)
}

// get value according to key
func (f *file) Get(key string) interface{} {
	b, err := ioutil.ReadFile(f.getFileName(key))
	if err != nil {
		return nil
	}
	return unmarshal(b)
}

// delete key
func (f *file) Delete(key string) error {
	return os.Remove(f.getFileName(key))
}

// expire session
func (f *file) Expire() error {
	return os.Remove(f.getDir())
}

// change mtime and atime
func (f *file) Update() error {
	return os.Chtimes(f.getDir(), time.Now(), time.Now())
}

// mtime of session directory
func (f *file) LastUpdate() time.Time {
	stat, err := os.Stat(f.getDir())
	if err != nil {
		return time.Time{}
	}
	return stat.ModTime()
}

// get session dir
func (f *file) getDir() string {
	return fmt.Sprintf("%v%v%v", f.path, f.pathSeparator, f.sid)
}

// get session file according to key
func (f *file) getFileName(key string) string {
	return fmt.Sprintf("%v%v%v", f.getDir(), f.pathSeparator, key)
}

// parse configuration
func (f *file) parseConf(conf string) {
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(conf), &m); err != nil {
		panic(err)
	}
	var ok bool
	f.path, ok = m["path"]
	if !ok {
		panic("has no path specified in conf")
	}
	f.pathSeparator, ok = m["separator"]
	if !ok {
		panic("has no separator in conf")
	}
}

func init() { Register(STORE_FILE, newFileStore) }
