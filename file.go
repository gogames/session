// file store
package session

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const permission = 0775

type file struct {
	root          string
	pathSeparator string
	generateID    func() string
}

var _ SessionStore = file{}

func NewFileStore(IDGenerator func() string, rootPath string, pathSeparator string) file {
	if err := os.MkdirAll(rootPath, permission); err != nil {
		panic(err)
	}
	if err := os.Chmod(rootPath, permission); err != nil {
		panic(err)
	}
	if IDGenerator == nil {
		IDGenerator = DefaultGenerator
	}
	return file{rootPath, pathSeparator, IDGenerator}
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
		if err := os.Mkdir(directory, permission); err != nil {
			log.Println(err)
			continue
		}
		if err := os.Chmod(directory, permission); err != nil {
			log.Println(err)
			continue
		}
		return id
	}
}

// set value
func (f file) Set(ID string, key string, val interface{}) error {
	if ID == "" {
		return nil
	}
	return ioutil.WriteFile(f.filePath(ID, key), marshal(val), permission)
}

// get value according to key
func (f file) Get(ID string, key string) interface{} {
	if ID == "" {
		return nil
	}
	b, err := ioutil.ReadFile(f.filePath(ID, key))
	if err != nil {
		return nil
	}
	return unmarshal(b)
}

// delete key
func (f file) Delete(ID string, key string) error {
	if ID == "" {
		return nil
	}
	return os.Remove(f.filePath(ID, key))
}

// expire session
func (f file) Expire(ID string) error {
	if ID == "" {
		return nil
	}
	return os.RemoveAll(f.directoryPath(ID))
}

// change mtime and atime
func (f file) Update(ID string) error {
	if ID == "" {
		return nil
	}
	t := time.Now()
	return os.Chtimes(f.directoryPath(ID), t, t)
}

// Flush remove all session
func (f file) Flush() error {
	return os.RemoveAll(f.root)
}

// GC removes all expired sessions
func (f file) GC(lifeTime time.Duration, t time.Time) {
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
			if info.ModTime().Add(lifeTime).Before(t) {
				return os.RemoveAll(path)
			}
			return nil
		})
}
