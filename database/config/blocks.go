package config

import (
	"database/sql"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/sirupsen/logrus"
)

type MBlock struct {
	db *sql.DB
}

func NewMBlockWith(mdb *sql.DB) *MBlock {
	return &MBlock{db: mdb}
}

func CurrentMBlock() *MBlock {
	return &MBlock{db: db}
}

func (m *MBlock) StoreBlock(cid cid.Cid) error {
	_, err := m.db.Exec(fmt.Sprintf("insert into block(cid) values('%s')", cid.String()))
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (m *MBlock) DelBlock(cid cid.Cid) {
	panic("not impl")
}

func (m *MBlock) ListAllBlock(b func(cid string)) {
	rows, err := m.db.Query("select * from block")
	if err != nil {
		logrus.Error(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var block string
		err = rows.Scan(&block)
		if err != nil {
			return
		}
		b(block)
	}
}

func StoreBlock(cid cid.Cid) error {
	return CurrentMBlock().StoreBlock(cid)
}
