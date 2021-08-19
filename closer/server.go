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

func CloseAll() {
	for name, close := range closer {
		err := close.Close()
		if err != nil {
			logrus.Errorf("close %s err %s", name, err)
		} else {
			logrus.Infof("close %s success", name)
		}
	}
}

type SimpleCloser struct {
	f func() error
}

func (s SimpleCloser) Close() error {
	if s.f != nil {
		return s.f()
	}
	return nil
}

func NewSimpleCloser(f func() error) SimpleCloser {
	return SimpleCloser{f: f}
}
