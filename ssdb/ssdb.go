package ssdb

import "github.com/gogames/session"

const STORE_SSDB = "ssdb"

// TODO: implement ssdb session provider
func init() { session.Register(STORE_SSDB, nil) }
