package closer

import (
	"errors"
	"github.com/sirupsen/logrus"
	"io"
)

var closer = make(map[string]io.Closer)

func RegisterCloser(name string, closer2 io.Closer) error {
	if _, b := closer[name]; b {
		return errors.New("exist")
	}
	closer[name] = closer2
	return nil
}

func Close() {
	for name, close := range closer {
		close.Close()
		logrus.Infof("close %s success", name)
	}
}
