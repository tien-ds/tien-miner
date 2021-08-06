package orbitdb

import (
	orbitdb "berty.tech/go-orbit-db"
	"berty.tech/go-orbit-db/accesscontroller"
	"berty.tech/go-orbit-db/iface"
	"berty.tech/go-orbit-db/stores/basestore"
	"context"
	"gitee.com/fast_api/api"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/ipds"
	"github.com/ds/depaas/persistence"
	"github.com/sirupsen/logrus"
)

type OrbitDB struct {
	kv iface.KeyValueStore
}

var orbit iface.OrbitDB

///orbitdb/bafyreihg347wiiup3v3vczlsxvdaodujtmnc2jpn742xm2ak6eb3qedeoa/data
func init() {

	persistence.InitDB(func(args map[string]string) persistence.IFdb {
		ctx := context.Background()
		dir := args["home_db"]
		//logger, _ := zap.NewDevelopment()
		orbit, err := orbitdb.NewOrbitDB(ctx, ipds.GetApi(), &orbitdb.NewOrbitDBOptions{
			Directory: &dir,
			//Logger: logger
		})
		checkError(err)
		logrus.Info(orbit.Identity().ID)
		ac := &accesscontroller.CreateAccessControllerOptions{
			Access: map[string][]string{
				"write": {
					orbit.Identity().ID,
				},
			},
		}

		kv, err := orbit.KeyValue(context.Background(), args["peer"], &iface.CreateDBOptions{
			AccessController: ac,
		})

		api.GET(func() interface{} {
			aController := kv.AccessController()
			return aController
		}, "/api/pool/kv")

		checkError(err)

		api.GET(func() interface{} {
			status := kv.ReplicationStatus()
			return map[string]interface{}{
				"progress": status.GetProgress(),
				"max":      status.GetMax(),
				"queue":    status.GetQueued(),
				"buffered": status.GetBuffered(),
			}
		}, "/api/db/status")

		api.GET(func() interface{} {
			return kv.Address().String()
		}, "/api/db/id")

		api.GET(func() interface{} {
			cid, _ := basestore.SaveSnapshot(ctx, kv)
			return map[string]string{
				"cid": cid.String(),
			}
		}, "/api/db/snapshot")

		dbName := args["peer"]
		//
		checkError(kv.Load(context.Background(), -1))
		//fmt.Println("LoadFromSnapshot", kv.LoadFromSnapshot(context.Background()))
		logrus.Info("load db success")
		if dbName == "data" {
			logrus.Info("load data from local")
			//c1, _ := cid.Decode("QmQFAinb8WvsvuqsN7jP5S2VR7RRgNZ4R3ozzw8YBskNdr")
			//kv.LoadMoreFrom(context.Background(), 0, []cid.Cid{c1})
		}
		logrus.Infof("db peer %s:", kv.Address().String())
		o := &OrbitDB{kv: kv}
		err = closer.RegisterCloser("OrbitDB", o)
		if err != nil {
			logrus.Error(err)
		}
		return o
	})

}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func (orb *OrbitDB) Close() error {
	err := orb.kv.Close()
	return err
}

func (orb *OrbitDB) Get(key []byte) ([]byte, error) {
	return orb.kv.Get(context.Background(), string(key))
}

func (orb *OrbitDB) Put(key []byte, value []byte) error {
	_, err := orb.kv.Put(context.Background(), string(key), value)
	return err
}

func (orb *OrbitDB) Delete(key []byte) error {
	_, err := orb.kv.Delete(context.Background(), string(key))
	return err
}

func (orb *OrbitDB) Count() (uint64, error) {
	return uint64(len(orb.kv.All())), nil
}

func (orb *OrbitDB) Range(filter persistence.KvFilter) {
	for key, values := range orb.kv.All() {
		if filter([]byte(key), values) {
			break
		}
	}
}

func (orb *OrbitDB) check() {

}
