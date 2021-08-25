package ipds

import (
	"encoding/json"
	"errors"
	"fmt"
	nconf "github.com/ds/depaas/database/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"io"
	"os"
	"path/filepath"
	"sync"

	filestore "github.com/ipfs/go-filestore"
	keystore "github.com/ipfs/go-ipfs/keystore"
	repo "github.com/ipfs/go-ipfs/repo"
	dir "github.com/ipfs/go-ipfs/thirdparty/dir"

	ds "github.com/ipfs/go-datastore"
	measure "github.com/ipfs/go-ds-measure"
	lockfile "github.com/ipfs/go-fs-lock"
	config "github.com/ipfs/go-ipfs-config"
	logging "github.com/ipfs/go-log"
	homedir "github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
)

const LockFile = "repo.lock"

var log = logging.Logger("repo")

// RepoVersion version number that we are currently expecting to see
var RepoVersion = 11

var migrationInstructions = `See https://github.com/ipfs/fs-repo-migrations/blob/master/run.md
Sorry for the inconvenience. In the future, these will run automatically.`

var programTooLowMessage = `Your programs version (%d) is lower than your repos (%d).
Please update ipfs to a version that supports the existing repo, or run
a migration in reverse.

See https://github.com/ipfs/fs-repo-migrations/blob/master/run.md for details.`

var (
	ErrNoVersion     = errors.New("no version file found, please run 0-to-1 migration tool.\n" + migrationInstructions)
	ErrOldRepo       = errors.New("ipfs repo found in old '~/.go-ipfs' location, please run migration tool.\n" + migrationInstructions)
	ErrNeedMigration = errors.New("ipfs repo needs migration")
)

type NoRepoError struct {
	Path string
}

var _ error = NoRepoError{}

func (err NoRepoError) Error() string {
	return fmt.Sprintf("no DS repo found in %s.\nplease run: 'ipfs init'", err.Path)
}

const apiFile = "api"

var (
	packageLock sync.Mutex

	onlyOne repo.OnlyOne
)

var swarm = `/key/swarm/psk/1.0.0/
/base16/
5cfce0358c1b154189c9921414104894307fa2d3da75ee1cd7404bbf8c905b7d

`

// PriRepo represents an IPFS FileSystem Repo. It is safe for use by multiple
// callers.
type PriRepo struct {
	// has Close been called already
	closed bool
	// path is the file-system path
	path string
	// lockfile is the file system lock to prevent others from opening
	// the same fsrepo path concurrently
	lockfile  io.Closer
	configDir string
	config    *config.Config
	ds        repo.Datastore
	keystore  keystore.Keystore
	filemgr   *filestore.FileManager
}

var _ repo.Repo = (*PriRepo)(nil)

func OpenRepo(repoPath string, configDir string) (repo.Repo, error) {
	return open(repoPath, configDir)
}

func open(repoPath string, configDir string) (repo.Repo, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	r, err := newFSRepo(repoPath, configDir)
	if err != nil {
		return nil, err
	}

	// Check if its initialized
	if err := checkInitialized(r.path); err != nil {
		return nil, err
	}

	r.lockfile, err = lockfile.Lock(r.configDir, LockFile)
	if err != nil {
		return nil, err
	}
	keepLocked := false
	defer func() {
		// unlock on error, leave it locked on success
		if !keepLocked {
			r.lockfile.Close()
		}
	}()

	//// Check version, and error out if not matching
	//ver, err := mfsr.RepoPath(r.path).Version()
	//if err != nil {
	//	if os.IsNotExist(err) {
	//		return nil, ErrNoVersion
	//	}
	//	return nil, err
	//}
	//
	//if RepoVersion > ver {
	//	return nil, ErrNeedMigration
	//} else if ver > RepoVersion {
	//	// program version too low for existing repo
	//	return nil, fmt.Errorf(programTooLowMessage, RepoVersion, ver)
	//}

	// check repo path, then check all constituent parts.
	if err := dir.Writable(r.path); err != nil {
		return nil, err
	}

	if err := r.openConfig(); err != nil {
		return nil, err
	}

	if err := r.openDatastore(); err != nil {
		return nil, err
	}

	//if err := r.openKeystore(); err != nil {
	//	return nil, err
	//}

	if r.config.Experimental.FilestoreEnabled || r.config.Experimental.UrlstoreEnabled {
		r.filemgr = filestore.NewFileManager(r.ds, filepath.Dir(r.path))
		r.filemgr.AllowFiles = r.config.Experimental.FilestoreEnabled
		r.filemgr.AllowUrls = r.config.Experimental.UrlstoreEnabled
	}

	keepLocked = true
	return r, nil
}

func newFSRepo(rpath string, configDir string) (*PriRepo, error) {
	expPath, err := homedir.Expand(filepath.Clean(rpath))
	if err != nil {
		return nil, err
	}

	return &PriRepo{path: expPath, configDir: configDir}, nil
}

func checkInitialized(path string) error {
	if !isInitializedUnsynced() {
		return NoRepoError{Path: path}
	}
	return nil
}

func configIsInitialized() bool {
	return nconf.GetDSConfig() != ""
}

func genConfig(conf *config.Config) error {
	if configIsInitialized() {
		return nil
	}
	bytes, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	return nconf.SetDSConfig(string(bytes))
}

func DSInitConfig(conf *config.Config) error {

	packageLock.Lock()
	defer packageLock.Unlock()

	if isInitializedUnsynced() {
		return nil
	}

	if err := genConfig(conf); err != nil {
		return err
	}

	//if err := initSpec(repoPath, conf.Datastore.Spec); err != nil {
	//	return err
	//}

	//if err := mfsr.RepoPath(repoPath).WriteVersion(RepoVersion); err != nil {
	//	return err
	//}

	return nil
}

func LockedByOtherProcess(repoPath string) (bool, error) {
	repoPath = filepath.Clean(repoPath)
	locked, err := lockfile.Locked(repoPath, LockFile)
	if locked {
		log.Debugf("(%t)<->Lock is held at %s", locked, repoPath)
	}
	return locked, err
}

func (r *PriRepo) Keystore() keystore.Keystore {
	return r.keystore
}

func (r *PriRepo) Path() string {
	return r.path
}

// SetAPIAddr writes the API Addr to the /api file.
func (r *PriRepo) SetAPIAddr(addr ma.Multiaddr) error {
	panic("no impl")
}

// openConfig returns an error if the config file is not present.
func (r *PriRepo) openConfig() error {
	dsConfig := nconf.GetDSConfig()
	var dsConf config.Config
	err := json.Unmarshal([]byte(dsConfig), &dsConf)
	if err == nil {
		r.config = &dsConf
	}
	return err
}

func (r *PriRepo) openKeystore() error {
	ksp := filepath.Join(r.path, "keystore")
	ks, err := keystore.NewFSKeystore(ksp)
	if err != nil {
		return err
	}

	r.keystore = ks

	return nil
}

// openDatastore returns an error if the config file is not present.
func (r *PriRepo) openDatastore() error {
	if r.config.Datastore.Type != "" || r.config.Datastore.Path != "" {
		return fmt.Errorf("old style datatstore config detected")
	} else if r.config.Datastore.Spec == nil {
		return fmt.Errorf("required Datastore.Spec entry missing from config file")
	}
	if r.config.Datastore.NoSync {
		log.Warn("NoSync is now deprecated in favor of datastore specific settings. If you want to disable fsync on flatfs set 'sync' to false. See https://github.com/ipfs/go-ipfs/blob/master/docs/datastores.md#flatfs.")
	}

	dsc, err := fsrepo.AnyDatastoreConfig(r.config.Datastore.Spec)
	if err != nil {
		return err
	}

	//r.path mul path
	d, err := dsc.Create(r.path)
	if err != nil {
		return err
	}
	r.ds = d

	// Wrap it with metrics gathering
	prefix := "ipfs.fsrepo.datastore"
	r.ds = measure.New(prefix, r.ds)

	return nil
}

// Close closes the PriRepo, releasing held resources.
func (r *PriRepo) Close() error {
	packageLock.Lock()
	defer packageLock.Unlock()

	if r.closed {
		return errors.New("repo is closed")
	}

	err := os.Remove(filepath.Join(r.path, apiFile))
	if err != nil && !os.IsNotExist(err) {
		log.Warn("error removing api file: ", err)
	}

	if err := r.ds.Close(); err != nil {
		return err
	}
	r.closed = true
	return r.lockfile.Close()
}

// Config the current config. This function DOES NOT copy the config. The caller
// MUST NOT modify it without first calling `Clone`.
//
// Result when not Open is undefined. The method may panic if it pleases.
func (r *PriRepo) Config() (*config.Config, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	if r.closed {
		return nil, errors.New("cannot access config, repo not open")
	}
	return r.config, nil
}

func (r *PriRepo) FileManager() *filestore.FileManager {
	return r.filemgr
}

func (r *PriRepo) BackupConfig(prefix string) (string, error) {
	panic("no impl")
}

// SetConfig updates the PriRepo's config. The user must not modify the config
// object after calling this method.
func (r *PriRepo) SetConfig(updated *config.Config) error {
	bytes, err := json.Marshal(updated)
	if err != nil {
		return err
	}
	return nconf.SetDSConfig(string(bytes))
}

// GetConfigKey retrieves only the value of a particular key.
func (r *PriRepo) GetConfigKey(key string) (interface{}, error) {
	panic("no impl")
}

// SetConfigKey writes the value of a particular key.
func (r *PriRepo) SetConfigKey(key string, value interface{}) error {
	panic("no impl")
}

// Datastore returns a repo-owned datastore. If PriRepo is Closed, return value
// is undefined.
func (r *PriRepo) Datastore() repo.Datastore {
	packageLock.Lock()
	d := r.ds
	packageLock.Unlock()
	return d
}

// GetStorageUsage computes the storage space taken by the repo in bytes
func (r *PriRepo) GetStorageUsage() (uint64, error) {
	return ds.DiskUsage(r.Datastore())
}

func (r *PriRepo) SwarmKey() ([]byte, error) {
	return []byte(swarm), nil
}

var _ io.Closer = &PriRepo{}
var _ repo.Repo = &PriRepo{}

// IsInitialized returns true if the repo is initialized at provided |path|.
func IsInitialized() bool {
	// packageLock is held to ensure that another caller doesn't attempt to
	// GenConfig or Remove the repo while this call is in progress.
	packageLock.Lock()
	defer packageLock.Unlock()

	return isInitializedUnsynced()
}

// private methods below this point. NB: packageLock must held by caller.

// isInitializedUnsynced reports whether the repo is initialized. Caller must
// hold the packageLock.
func isInitializedUnsynced() bool {
	return configIsInitialized()
}
