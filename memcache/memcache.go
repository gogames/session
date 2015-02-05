package memcache

import "github.com/gogames/session"

const STORE_MEMCACHE = "memcache"

// TODO: implement memcache session provider
func init() { session.Register(STORE_MEMCACHE, nil) }
