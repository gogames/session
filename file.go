// file store
package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/hprose/hprose-go/hprose"
)

const STORE_FILE = "file"

type file struct {
	sid           string
	path          string
	pathSeparator string
}

func init() {
	Register(STORE_FILE, newFileStore)
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

// serialize
func (f *file) marshal(data interface{}) []byte {
	b, err := hprose.Marshal(data)
	if err != nil {
		panic(err) // fail quick, in case of those structs can not be hashed
	}
	return b
}

// unserialize
func (f *file) unmarshal(b []byte) interface{} {
	var v interface{}
	err := hprose.Unmarshal(b, &v)
	if err != nil {
		panic(err) // fail quick, in case of those structs can not be hashed
	}
	return v
}

func (f *file) getDir() string {
	return fmt.Sprintf("%v%v%v", f.path, f.pathSeparator, f.sid)
}

func (f *file) getFileName(key interface{}) string {
	return fmt.Sprintf("%v%v%v", f.getDir(), f.pathSeparator, key)
}

func (f *file) SessionId() string {
	return f.sid
}

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

func (f *file) Init() {
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		if err = os.Mkdir(f.path, os.ModePerm); err != nil {
			panic(fmt.Errorf("can not mkdir: %v", err))
		}
	}
}

func (f *file) Iterate(fc func(key, val interface{})) {
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
		fc(info.Name(), f.unmarshal(b))
		return nil
	})
}

func (f *file) Set(key, val interface{}) error {
	return ioutil.WriteFile(f.getFileName(key), f.marshal(val), os.ModePerm)
}

func (f *file) Get(key interface{}) interface{} {
	b, err := ioutil.ReadFile(f.getFileName(key))
	if err != nil {
		return nil
	}
	return f.unmarshal(b)
}

func (f *file) Delete(key interface{}) error {
	return os.Remove(f.getFileName(key))
}

func (f *file) Expire() error {
	return os.Remove(f.getDir())
}

func (f *file) Update() error {
	return os.Chtimes(f.getDir(), time.Now(), time.Now())
}

func (f *file) LastUpdate() time.Time {
	stat, err := os.Stat(f.getDir())
	if err != nil {
		return time.Time{}
	}
	return stat.ModTime()
}
