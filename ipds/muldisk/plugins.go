package muldisk

import (
	"fmt"
	"github.com/dustin/go-humanize"
	badgerds "github.com/ipfs/go-ds-badger"
	"github.com/ipfs/go-ipfs/plugin"
	"github.com/ipfs/go-ipfs/repo"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

// Plugins is exported list of plugins that will be loaded
var Plugins = []plugin.Plugin{
	&mulDataStorePlugin{},
}

type mulDataStorePlugin struct{}

var _ plugin.PluginDatastore = (*mulDataStorePlugin)(nil)

func (*mulDataStorePlugin) Name() string {
	return "ds-mul-datastore"
}

func (*mulDataStorePlugin) Version() string {
	return "0.1.0"
}

func (*mulDataStorePlugin) Init(_ *plugin.Environment) error {
	return nil
}

func (*mulDataStorePlugin) DatastoreTypeName() string {
	return "ds-mul-datastore"
}

type datastoreConfig struct {
	path       string
	syncWrites bool
	truncate   bool

	vlogFileSize int64
}

// DatastoreConfigParser BadgerdsDatastoreConfig returns a configuration stub for a badger datastore
// from the given parameters
func (*mulDataStorePlugin) DatastoreConfigParser() fsrepo.ConfigFromMap {
	return func(params map[string]interface{}) (fsrepo.DatastoreConfig, error) {
		var c datastoreConfig
		var ok bool

		c.path, ok = params["path"].(string)
		if !ok {
			return nil, fmt.Errorf("'path' field is missing or not string")
		}

		sw, ok := params["syncWrites"]
		if !ok {
			c.syncWrites = false
		} else {
			if swb, ok := sw.(bool); ok {
				c.syncWrites = swb
			} else {
				return nil, fmt.Errorf("'syncWrites' field was not a boolean")
			}
		}

		truncate, ok := params["truncate"]
		if !ok {
			c.truncate = true
		} else {
			if truncate, ok := truncate.(bool); ok {
				c.truncate = truncate
			} else {
				return nil, fmt.Errorf("'truncate' field was not a boolean")
			}
		}

		vls, ok := params["vlogFileSize"]
		if !ok {
			// default to 1GiB
			c.vlogFileSize = badgerds.DefaultOptions.ValueLogFileSize
		} else {
			if vlogSize, ok := vls.(string); ok {
				s, err := humanize.ParseBytes(vlogSize)
				if err != nil {
					return nil, err
				}
				c.vlogFileSize = int64(s)
			} else {
				return nil, fmt.Errorf("'vlogFileSize' field was not a string")
			}
		}

		return &c, nil
	}
}

func (c *datastoreConfig) DiskSpec() fsrepo.DiskSpec {
	return map[string]interface{}{
		"type": "badgerds",
		"path": c.path,
	}
}

func (c *datastoreConfig) Create(path string) (repo.Datastore, error) {

	defopts := DefaultOptions
	defopts.SyncWrites = c.syncWrites
	defopts.Truncate = c.truncate
	defopts.ValueLogFileSize = c.vlogFileSize

	return NewMulDataStore(&defopts, path)
}
