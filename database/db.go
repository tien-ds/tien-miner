package database

import (
	"database/sql"
	"github.com/ds/depaas/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"path"
)

var appDb string

var initSQL = `create table  if not exists app (key text not null primary key,value text);
create table  if not exists block (cid text not null primary key)`

func GetDBPath() string {
	initDatabase()
	return appDb
}

func initDatabase() {
	dir := utils.GetConfigDir()
	appDb = path.Join(dir, "app.db")
	if !utils.Exist(appDb) {
		logrus.Tracef("create db %s", appDb)
		db, err := sql.Open("sqlite3", appDb)
		if err != nil {
			logrus.Fatalf("db %s %s", appDb, err)
		}
		defer db.Close()
		_, err = db.Exec(initSQL)
		if err != nil {
			logrus.Errorf("%q: %s", err, initSQL)
			return
		}
		logrus.Debug("db init ok")
	}
}

func OpenDb() (*sql.DB, error) {
	return sql.Open("sqlite3", GetDBPath())
}

func OpenDBFile(name string) (*sql.DB, error) {
	initDatabase()
	return sql.Open("sqlite3", name)
}
