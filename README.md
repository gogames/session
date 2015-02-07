session provider
===
[![GoDoc](http://godoc.org/github.com/gogames/session?status.svg)](http://godoc.org/github.com/gogames/session)
[![Build Status](https://travis-ci.org/gogames/session.svg?branch=master)](https://travis-ci.org/gogames/session)
[![status](https://sourcegraph.com/api/repos/github.com/gogames/session/.badges/status.png)](https://sourcegraph.com/github.com/gogames/session)

### Features

* Stand alone, not coupled with cookie

* Thread safe

* LRU for quick gc scan

### Install

``` go get -u github.com/gogames/session ```


### How to use

```go
	gcFrequency := time.Second
	sessionLifeTime := time.Hour
	
	sess := session.NewSession(session.STORE_FILE, gcFrequency, sessionLifeTime, `{"path":"session_path", "separator": "/"}`)

	// create a new session
	sid, err := sess.Set("", "key", "value")

	// get
	val := sess.Get(sid, "key")

	// set
	_, err = sess.Set(sid, "key2", "value2")

	// delete 
	err = sess.Delete(sid, "key")

	// update session life time
	err = sess.Update(sid)

	// expire session
	err = sess.Expire(sid)

	// close session, wait util all session operations done
	// useful function used in graceful exit program
	sess.Close()
```

### TODO

Implement the following session providers

* Redis
* SSDB
* Memcache
* Mongo
* Mysql
* Postgresql
* Sqlite3
