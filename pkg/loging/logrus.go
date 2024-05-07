package loging

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

type hook struct {
	Writers   []io.Writer
	LogLevels []logrus.Level
}

func (h *hook) Fire(e *logrus.Entry) error {

	l, err := e.Bytes()

	if err != nil {
		return err
	}

	for _, v := range h.Writers {
		v.Write(l)
	}

	return nil

}

func (h *hook) Levels() []logrus.Level {
	return h.LogLevels
}

func InitLogrus(logFilePath string) (*logrus.Logger, error) {

	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			fileName := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%s", fileName, f.Line)
		},

		FullTimestamp: true,
	}

	h := &hook{
		Writers:   []io.Writer{},
		LogLevels: logrus.AllLevels,
	}

	if logFilePath != "" {
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		h.Writers = append(h.Writers, logFile)
	}

	l.SetOutput(io.Discard)

	h.Writers = append(h.Writers, os.Stdout)

	l.AddHook(h)

	l.SetLevel(logrus.TraceLevel)

	return l, nil

}
