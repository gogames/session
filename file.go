// file store
package session

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type file struct {
	root          string
	pathSeparator string
	generateID    func() string
	lifeTime      time.Duration
}

var _ SessionStore = file{}

func NewFileStore(IDGenerator func() string, rootPath string, pathSeparator string, lifeTime time.Duration) file {
	if err := os.MkdirAll(rootPath, os.ModePerm); err != nil {
		panic(err)
	}
	if IDGenerator == nil {
		IDGenerator = DefaultGenerator
	}
	return file{rootPath, pathSeparator, IDGenerator, lifeTime}
}

func (f file) directoryPath(ID string) string {
	return strings.Join([]string{
		strings.TrimRight(f.root, f.pathSeparator),
		strings.TrimLeft(ID, f.pathSeparator),
	}, f.pathSeparator)
}

func (f file) filePath(ID, key string) string {
	return strings.Join([]string{
		strings.TrimRight(f.root, f.pathSeparator),
		strings.TrimLeft(ID+f.pathSeparator+key, f.pathSeparator),
	}, f.pathSeparator)
}

func (f file) GenerateID() string {
	for {
		id := f.generateID()
		directory := f.directoryPath(id)
		if err := os.Mkdir(directory, os.ModePerm); err != nil {
			continue
		}
		return id
	}
}

// set value
func (f file) Set(ID string, key string, val interface{}) error {
	return ioutil.WriteFile(f.filePath(ID, key), marshal(val), os.ModePerm)
}

// get value according to key
func (f file) Get(ID string, key string) interface{} {
	b, err := ioutil.ReadFile(f.filePath(ID, key))
	if err != nil {
		return nil
	}
	return unmarshal(b)
}

// delete key
func (f file) Delete(ID string, key string) error {
	return os.Remove(f.filePath(ID, key))
}

// expire session
func (f file) Expire(ID string) error {
	return os.RemoveAll(f.directoryPath(ID))
}

// change mtime and atime
func (f file) Update(ID string) error {
	t := time.Now()
	return os.Chtimes(f.directoryPath(ID), t, t)
}

// Flush remove all session
func (f file) Flush() error {
	return os.RemoveAll(f.root)
}

// mtime of session directory
func (f file) lastUpdate(ID string) time.Time {
	stat, err := os.Stat(f.directoryPath(ID))
	if err != nil {
		return time.Time{}
	}
	return stat.ModTime()
}

// GC removes all expired sessions
func (f file) GC(t time.Time) {
	filepath.Walk(f.root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			if path == f.root {
				return nil
			}
			if info.ModTime().Add(f.lifeTime).Before(t) {
				return os.RemoveAll(path)
			}
			return nil
		})
}
