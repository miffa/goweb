package bufferpool

import (
	"bytes"
	"sync"
)

var Pool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		return new(bytes.Buffer)
	},
}

//
func Get() *bytes.Buffer {
	return Pool.Get().(*bytes.Buffer)
}

func Put(b *bytes.Buffer) {
	Pool.Put(b)
}
