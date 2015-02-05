package session

import (
	"bytes"
	"encoding/gob"
)

const _KEY = "data"

func init() { gob.Register(map[string]interface{}{}) }

func marshal(d interface{}) []byte {
	data := map[string]interface{}{_KEY: d}
	gob.Register(d)
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(data); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func unmarshal(b []byte) interface{} {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	var v = make(map[string]interface{})
	if err := dec.Decode(&v); err != nil {
		panic(err)
	}
	return v[_KEY]
}
