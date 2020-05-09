package seqid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"os"
	"sync/atomic"
	"time"
)

// objectIDCounter is atomically incremented when generating a new ObjectId
// using NewObjectId() function. It's used as a counter part of an id.
// This id is initialized with a random value.
var objectIDCounter = randInt()

// machineId stores machine id generated once and used in subsequent calls
// to NewObjectId function.
var machineID = readMachineID()

// randInt generates a random uint32
func randInt() uint32 {
	b := make([]byte, 3)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(fmt.Errorf("Cannot generate random number: %v;", err))
	}
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

// readMachineId generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, it will cause
// a runtime error.
func readMachineID() []byte {
	id := make([]byte, 3)
	if hostname, err := os.Hostname(); err == nil {
		hw := fnv.New32a()
		hw.Write([]byte(hostname))
		copy(id, hw.Sum(nil))
	} else {
		// Fallback to rand number if machine id can't be gathered
		if _, randErr := rand.Reader.Read(id); randErr != nil {
			panic(fmt.Errorf("Cannot get hostname nor generate a random number: %v; %v", err, randErr))
		}
	}
	return id
}

// New generates a globaly unique ID
func New() uint64 {
	var id [8]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(id[:], uint32(time.Now().Unix()))
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIDCounter, 1)
	id[4] = byte(i >> 16)
	id[5] = byte(i >> 8)
	id[6] = byte(i)
	// Machine, first byte of md5(hostname)
	id[7] = machineID[0]
	return binary.BigEndian.Uint64(id[:])
}

func MachineID1stByte() byte {
	return machineID[0]
}

// Time returns the timestamp part of the id.
// It's a runtime error to call this method with an invalid id.
func ToTime(seqId uint64) time.Time {
	// First 4 bytes of ObjectId is 32-bit big-endian seconds from epoch.
	return time.Unix(int64(seqId>>32), 0)
}

func ToCounter(seqId uint64) (cnt uint32) {
	cnt = uint32(seqId)
	cnt = cnt >> 8
	return
}

func ToMachineId(seqId uint64) uint8 {
	return uint8(seqId)
}

func TimeString() string {
	return time.Now().Format(time.RFC3339)
}
