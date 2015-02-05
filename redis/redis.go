package redis

import "github.com/gogames/session"

const STORE_REIDS = "redis"

// TODO: implement redis session provider
func init() { session.Register(STORE_REIDS, nil) }
