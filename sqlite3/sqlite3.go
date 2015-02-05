package sqlite3

import "github.com/gogames/session"

const STORE_SQLITE3 = "sqlite3"

// TODO: implement sqlite3 session provider
func init() { session.Register(STORE_SQLITE3, nil) }
