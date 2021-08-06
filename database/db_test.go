package database

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestDb(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	create table foo (id integer not null primary key, name text);
	delete from foo;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		logrus.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		logrus.Fatal(err)
	}
	stmt, err := tx.Prepare("replace into foo(id, name) values(?, ?)")
	if err != nil {
		logrus.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("こんにちわ世界%03d", i))
		if err != nil {
			logrus.Fatal(err)
		}
	}
	tx.Commit()

	rows, err := db.Query("select id, name from foo")
	if err != nil {
		logrus.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			logrus.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		logrus.Fatal(err)
	}

	stmt, err = db.Prepare("select name from foo where id = ?")
	if err != nil {
		logrus.Fatal(err)
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow("3").Scan(&name)
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Println(name)

	_, err = db.Exec("delete from foo")
	if err != nil {
		logrus.Fatal(err)
	}

	_, err = db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
	if err != nil {
		logrus.Fatal(err)
	}

	rows, err = db.Query("select id, name from foo")
	if err != nil {
		logrus.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			logrus.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		logrus.Fatal(err)
	}
}
