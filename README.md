session provider
===
[![GoDoc](http://godoc.org/github.com/gogames/session?status.svg)](http://godoc.org/github.com/gogames/session)
[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/gogames/session)

### Features

* Stand alone, not coupled with cookie which is annoying...

* Thread safe


### Install

``` go get -u github.com/gogames/session ```


### How to use

```go
	sess := session.NewSession(session.STORE_FILE, time.Hour, `{"path":"session_path", "separator": "/"}`)

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
