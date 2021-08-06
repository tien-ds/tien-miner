package config

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/database"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

var db *sql.DB
var one sync.Once

func Init() {
	mdb, err := database.OpenDb()
	if err != nil {
		logrus.Error(err)
		os.Exit(0)
	}
	db = mdb
	err = closer.RegisterCloser("config.db", db)
	if err != nil {
		logrus.Error(err)
	}
}

type AppSetting struct {
	db *sql.DB
}

func NewAppSetting() *AppSetting {
	one.Do(Init)
	return &AppSetting{db: db}
}

func (a *AppSetting) SetConfig(key, value string) error {
	if value == "" {
		return errors.New("value empty")
	}
	_, err := a.db.Exec(fmt.Sprintf("replace into app(key, value) values('%s', '%s')", key, value))
	return err
}

func (a *AppSetting) GetConfig(key string) (string, error) {
	rows, err := a.db.Query(fmt.Sprintf("select key,value from app where key='%s'", key))
	if err != nil {
		logrus.Fatal(err)
		return "", err
	}
	defer rows.Close()
	nRow := 0
	for rows.Next() {
		nRow++
		var key, value string
		err = rows.Scan(&key, &value)
		if err != nil {
			logrus.Fatal(err)
			return "", err
		}
		if value != "" {
			return value, nil
		}
	}
	if nRow == 0 {
		return "", errors.New("db empty")
	}
	err = rows.Err()
	if err != nil {
		logrus.Fatal(err)
	}
	return "", err
}

func SetConfigKey(key string, value string) error {
	return NewAppSetting().SetConfig(key, value)
}

func GetConfigKey(key string) (string, error) {
	return NewAppSetting().GetConfig(key)
}
