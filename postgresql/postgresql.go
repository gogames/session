package postgresql

import "github.com/gogames/session"

const STORE_POSTGRESQL = "postgresql"

// TODO: implement postgresql session provider
func init() { session.Register(STORE_POSTGRESQL, nil) }
