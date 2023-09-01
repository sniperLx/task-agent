package log

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
	once   sync.Once
)

func init() {
	InitLogger()
}

func InitLogger() *logrus.Logger {
	once.Do(func() {
		logger = logrus.New()
		logger.SetFormatter(&MyFormatter{})
		logger.SetOutput(os.Stdout)
		logger.SetLevel(logrus.WarnLevel)
		logger.SetReportCaller(true)
	})
	return logger
}

type MyFormatter struct{}

func (f *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var file string
	var line int
	if entry.Caller != nil {
		file = filepath.Base(entry.Caller.File)
		line = entry.Caller.Line
	}
	timestamp := entry.Time.Format("2023-09-01 22:00:00")
	msg := fmt.Sprintf("%s [%s] [%s:%d] [goid:%d] %s\n", timestamp, strings.ToUpper(entry.Level.String()), file, line,
		getGId(), entry.Message)

	return []byte(msg), nil
}

func getGId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	id, _ := strconv.ParseUint(string(b), 10, 64)
	return id
}
