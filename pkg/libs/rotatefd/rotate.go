package rotatefd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	Byte = 1
	KiB  = 1024 * Byte
	MiB  = 1024 * KiB
	GiB  = 1024 * MiB
)

var (
	ErrorNilFD    = errors.New("underlying fd in rotateFile is nil")
	ErrorFullDisk = errors.New("underlying fd in rotateFile is full")
)

type rotateFile struct {
	date        time.Time // will rotate after this day ends
	maxFragSize uint64    // will rotate after file size exceeds this
	curFragSize uint64
	fragidx     uint32
	basePath    string // will rotate as basePath.date.fragidx
	fd          *os.File
	wfd         chan []byte
}

func NewRotateFile(basePath string, maxFragSize uint64) io.WriteCloser {
	if basePath == "" {
		return os.Stdout
	}
	var rf = &rotateFile{
		date:        time.Now(),
		maxFragSize: maxFragSize,
		fragidx:     0,
		basePath:    basePath,
		fd:          nil,
	}
	rf.wfd = make(chan []byte, 2048)
	rf.rotate(false)
	go rf.writeLoop()
	return rf
}

func (self *rotateFile) dateStr() string {
	return fmt.Sprintf("%d-%02d-%02d", self.date.Year(), self.date.Month(), self.date.Day())
}

func (self *rotateFile) archive() {
	var rotatePath = fmt.Sprintf("%s.%s.%03d", self.basePath, self.dateStr(), self.fragidx)
	if err := os.Rename(self.basePath, rotatePath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rename %s to %s: %v\n", self.basePath, rotatePath, err)
	}
}

func (self *rotateFile) createfile() {
	var err error
	if self.fd, err = os.OpenFile(self.basePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0666); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to OpenFile(\"%s\", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666): %v\n", self.basePath, err)
		os.Exit(1)
	}
}

func (self *rotateFile) scanToNextIdx() {
	datePath := fmt.Sprintf("%s.%s.", self.basePath, self.dateStr())
	if existing, err := filepath.Glob(datePath + "*"); err == nil {
		for _, rotatePath := range existing {
			if idx, err := strconv.ParseUint(rotatePath[len(datePath):], 10, 32); err == nil {
				if uint32(idx) >= self.fragidx {
					self.fragidx = uint32(idx) + 1
				}
			}
		}
	}
}

func (self *rotateFile) rotate(anotherday bool) {
	if self.fd == nil {
		// application initiallization, archive old file(if exists and have content), create new file
		if info, err := os.Stat(self.basePath); err == nil && info.Size() > 0 {
			self.date = info.ModTime()
			self.scanToNextIdx() // to avoid conflict with that day's existing archives
			self.archive()
			self.date = time.Now()
			self.fragidx = 0
			self.scanToNextIdx() // to avoid conflict with today's existing archives
		}
	} else {
		// normal rotate, archive current file, create new file
		self.fd.Close()
		self.fd = nil
		self.archive()
		if anotherday {
			self.date = time.Now()
			self.fragidx = 0
		} else {
			self.fragidx += 1
		}

	}
	self.createfile()
	self.curFragSize = 0
}

func (r *rotateFile) Close() error {
	close(r.wfd)
	if fd := r.fd; fd != nil {
		r.fd = nil
		return fd.Close()
	}
	return nil
}

func (r *rotateFile) Write(b []byte) (n int, err error) {
	select {
	case r.wfd <- b:
		return len(b), nil
	default:
		return 0, ErrorFullDisk
	}
	return
}

func (r *rotateFile) write(b []byte) (n int, err error) {
	if r.fd == nil {
		return 0, ErrorNilFD
	}
	var now = time.Now()
	var anotherday = now.YearDay() != r.date.YearDay() || now.Year() != r.date.Year()
	if r.curFragSize > r.maxFragSize || anotherday {
		r.rotate(anotherday)
	}

	if r.fd == nil {
		return 0, ErrorNilFD
	}
	n, err = r.fd.Write(b)
	r.curFragSize += uint64(n)
	return
}

func (r *rotateFile) writeLoop() {
	for {
		select {
		case b, ok := <-r.wfd:
			if !ok {
				return
			}
			r.write(b)
		}
	}
	return

}
