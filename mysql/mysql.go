package mysql

import "github.com/gogames/session"

const STORE_MYSQL = "mysql"

// TODO: implement mysql session provider
func init() { session.Register(STORE_MYSQL, nil) }
