session provider
===

* Stand alone, not coupled with cookie.

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
