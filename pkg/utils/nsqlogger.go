package utils

import (
	"runtime"
	"strconv"

	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
)

type NSQLogger struct {
	Log *logrus.Logger
}

func (n *NSQLogger) Output(calldepth int, text string) error {
	_, file, line, _ := runtime.Caller(calldepth)
	n.Log.WithFields(logrus.Fields{
		"location": file + ":" + strconv.Itoa(line),
	}).Warn(text)
	return nil
}
