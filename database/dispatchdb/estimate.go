package dispatchdb

import (
	"database/sql"
	"gitee.com/fast_api/api"
	"gitee.com/fast_api/api/def"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/utils"
	"github.com/sirupsen/logrus"
	"path"
)

var sdb *sql.DB

func init() {
	var initSQL = `CREATE TABLE "dispatcher" (
	"id"	INTEGER NOT NULL UNIQUE,
	"cid"	TEXT NOT NULL,
	"peer"	TEXT NOT NULL,
	"size"	INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);`
	dir := utils.GetConfigDir()
	appDb := path.Join(dir, "dipatcher.db")
	db, err := sql.Open("sqlite3", appDb)
	if err != nil {
		logrus.Fatalf("db %s %s", appDb, err)
	}
	sdb = db
	_, err = db.Exec(initSQL)

	closer.RegisterCloser("dipatcher.db", closer.NewSimpleCloser(func() error {
		return sdb.Close()
	}))

	api.GET(func(peer def.StringReq) interface{} {
		err, u := EstimateRepo(peer.String())
		if err != nil {
			panic(err)
		}
		return map[string]uint64{"size": u}
	}, "/api/pool/repo")
}

func Store(cid, peer string, size uint64) error {
	_, err := sdb.Exec("INSERT INTO dispatcher(cid,peer,size) VALUES(?,?,?)", cid, peer, size)
	if err != nil {
		return err
	}
	return nil
}

func EstimateRepo(peer string) (error, uint64) {
	result, err := sdb.Query("SELECT sum(size) FROM dispatcher WHERE peer = ?", peer)
	var size uint64
	if result.Next() {
		result.Scan(&size)
	} else {
		return err, 0
	}
	return nil, size
}
