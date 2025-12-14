package internal

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestCacheMaxSize(t *testing.T) {
	eval := is.New(t)

	c := NewMemoryCache(30*time.Second, 10)

	err := c.Set("1", &Record{
		Body: []byte("12345"),
	})
	eval.NoErr(err)

	err = c.Set("1", &Record{
		Body: []byte("123456"),
	})
	eval.NoErr(err)

	err = c.Set("2", &Record{
		Body: []byte("1234"),
	})

	eval.True(err == ErrCacheFull)

	time.Sleep(30 * time.Second)

	//this is needed to remove expired record
	c.Get("1")

	err = c.Set("2", &Record{
		Body: []byte("1234"),
	})
	eval.NoErr(err)
}
