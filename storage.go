package handshake

import (
	"bytes"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// StorageEngine type for enum
type StorageEngine int

const (
	// BoltEngine is the first supported StorageEngine
	BoltEngine StorageEngine = iota
	// future implimentations will go here, wasm -> js connector?
)

const (
	// DefaultStorageEngine is used to set the storage engine if none is set in
	// storage options
	DefaultStorageEngine = BoltEngine
	// DefaultBoltFilePath is the default path and file name for BoltDB storage
	DefaultBoltFilePath = "handshake.boltdb"
	// DefaultTLB is the name of the top level bucket for BoltDB
	DefaultTLB = "handshake"
	// GlobalConfigKey is the key string for where global-config is stored
	GlobalConfigKey = "global-config"
)

// Storage is the primary interface for interacting with the KV store in handshake
type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	List(path string) ([]string, error)
	Close() error
}

// StorageOptions are used to pass in initialization settings
type StorageOptions struct {
	Engine   StorageEngine
	FilePath string
}

// NewStorage initiates a new storage Interface
func NewStorage(opts StorageOptions) (Storage, error) {
	switch opts.Engine {
	case BoltEngine:
		return NewBoltStorage(opts)
	default:
		return nil, errors.New("invalid engine type")
	}
}

// NewBoltStorage takes StorageOptions as an argument and returns a reference to a BoltDB
// based implimentation of the Storage interface.
func NewBoltStorage(opts StorageOptions) (Storage, error) {
	tlb := DefaultTLB
	fp := DefaultBoltFilePath
	if opts.FilePath != "" {
		fp = opts.FilePath
	}
	db, err := bolt.Open(fp, 0666, nil)
	if err != nil {
		return nil, err
	}

	// ensure that top level bucket exists
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(tlb)); err != nil {
			return fmt.Errorf("error creating bucket: %s", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// ensure that globalConfig exists, and if not, initialize GlobalConfig
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tlb))
		blob := b.Get([]byte(GlobalConfigKey))
		if blob == nil {
			return b.Put([]byte(GlobalConfigKey), NewGlobalConfig().ToJSON())
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return BoltStorage{DB: db, TLB: tlb}, nil
}

// BoltStorage is a struct that conforms to the Storage interface for using
// BoltDB. DB is a reference to a boltDB instance and TLB stands for "top level bucket"
type BoltStorage struct {
	DB  *bolt.DB
	TLB string
}

// Get takes a key string and returns a byte slice or error from a BoltStorage struct. Get
// get retruns and empty byte slice if no key is found and/or the byte slice is blank. An error
// is returned if the key invalid in formatting, it is too long, or there is an underlying issue
// with boltDB
func (s BoltStorage) Get(key string) (value []byte, err error) {
	err = s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.TLB))
		value = b.Get([]byte(key))
		return nil
	})
	return value, err
}

// Set takes a key string and value byte slice returns an error from a BoltStorage struct.
// Set treats both create and updates the same. Errors are returned if the key has invalid syntax
// and if key or value are too long.
func (s BoltStorage) Set(key string, value []byte) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.TLB))
		return b.Put([]byte(key), value)
	})
}

// Delete takes a key string and returns an error from a BoltStorage struct.
func (s BoltStorage) Delete(key string) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.TLB))
		return b.Delete([]byte(key))
	})
}

// List takes a path and returns a slice of key paths formatted as strings or an error.
func (s BoltStorage) List(path string) (keys []string, err error) {
	p := []byte(path)
	err = s.DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(s.TLB)).Cursor()
		for k, _ := c.Seek(p); k != nil && bytes.HasPrefix(k, p); k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		return nil
	})
	return keys, err
}

// Close is used to close the Bolt DB engine and returns an error
func (s BoltStorage) Close() error {
	return s.DB.Close()
}
