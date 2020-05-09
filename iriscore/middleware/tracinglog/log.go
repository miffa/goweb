package tracinglog

import (
	"io"

	log "goweb/iriscore/libs/logrus"
)

const (
	LOG_F_JSON = iota
	LOG_F_TEXT
)

type loggerOption struct {
	Filepath   string
	RotateSize uint64 // M
	Format     int
}

type LogOption func(*loggerOption)

type DeckLogger struct {
	*log.Logger
	rfAccess io.WriteCloser
	loption  loggerOption
}

func WithFilePath(f string) LogOption {
	return func(o *loggerOption) {
		o.Filepath = f
	}
}

func WithJson() LogOption {
	return func(o *loggerOption) {
		o.Format = LOG_F_JSON
	}
}
func WithText() LogOption {
	return func(o *loggerOption) {
		o.Format = LOG_F_TEXT
	}
}

func WithRotateSize(s uint64) LogOption {
	return func(o *loggerOption) {
		o.RotateSize = s
	}
}

func NewDeckLogger(path string) *DeckLogger {
	/*
		dl := &DeckLogger{}
		dl.rfAccess = log.NewRotateFile(path, 100*log.MiB)
		dl.Logger = log.New()
		//dl.Logger.Formatter = &log.JSONFormatter{}
		dl.Json()
		dl.Logger.Out = dl.rfAccess
		return dl
	*/
	return NewDeckLoggerWithOption(WithRotateSize(100), WithJson(), WithFilePath(path))
}

func NewDeckLoggerWithOption(opts ...LogOption) *DeckLogger {
	dl := &DeckLogger{}
	for _, optfunc := range opts {
		optfunc(&dl.loption)
	}

	if dl.loption.RotateSize == 0 {
		dl.loption.RotateSize = 100
	}
	dl.rfAccess = log.NewRotateFile(dl.loption.Filepath, dl.loption.RotateSize*log.MiB)
	dl.Logger = log.New()
	//dl.Logger.Formatter = &log.JSONFormatter{}
	switch dl.loption.Format {
	case LOG_F_JSON:
		dl.Json()
	case LOG_F_TEXT:
		dl.Text()
	default:
		dl.Text()
	}

	dl.Logger.Out = dl.rfAccess
	return dl
}

func (dl *DeckLogger) Json() *DeckLogger {
	dl.Logger.Formatter = &log.JSONFormatter{}
	return dl
}

func (dl *DeckLogger) Text() *DeckLogger {
	dl.Logger.Formatter = &log.TextFormatter{}
	return dl
}

func (dl *DeckLogger) Close() error {
	return dl.rfAccess.Close()
}
