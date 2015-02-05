package mongo

import "github.com/gogames/session"

const STORE_MONGO = "mongo"

// TODO: implement mongo session provider
func init() { session.Register(STORE_MONGO, nil) }
