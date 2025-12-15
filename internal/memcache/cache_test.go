package memcache

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestCacheMaxSize(t *testing.T) {
	eval := is.New(t)

	c := NewMemoryCache(30*time.Second, 100, 10)

	c.Set("1", &Record{
		Body: []byte("12345"),
	})
	eval.Equal(c.Count(), 1)
	eval.Equal(c.remainingCapacity, 55)

	c.Set("1", &Record{
		Body: []byte("123456"),
	})
	eval.Equal(c.Count(), 1)
	eval.Equal(c.remainingCapacity, 54)

	c.Set("2", &Record{
		Body: []byte("1234"),
	})

	eval.Equal(c.Count(), 2)
	eval.Equal(c.remainingCapacity, 10)

	//one record will be removed as capacity is 10
	c.Set("3", &Record{
		Body: []byte("123456"),
	})

	eval.Equal(c.Count(), 2)
	eval.Equal(c.remainingCapacity, 10)
}

func TestCacheMaxRecordSize(t *testing.T) {
	eval := is.New(t)

	c := NewMemoryCache(30*time.Second, 100, 10)

	err := c.Set("1", &Record{
		Body: []byte("1234567890"),
	})
	eval.Equal(err, ErrMaxRecordSizeExceed)

	c = NewMemoryCache(30*time.Second, 100, 100)

	err = c.Set("1", &Record{
		Body: []byte("1234567890"),
	})
	eval.NoErr(err)
}
