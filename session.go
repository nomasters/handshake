package handshake

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	// DefaultSessionTTL is the default TTL before a Session closes
	DefaultSessionTTL = 15 * 60 // 15 minutes in seconds
	// DefaultMaxLoginAttempts is the number of times failed login attempts are allowed
	DefaultMaxLoginAttempts = 10
)

// Session is the primary struct for a logged in  user. It holds the profile data
// as well as settings information
type Session struct {
	Profile      Profile
	Storage      Storage
	Cipher       Cipher
	TTL          int64
	StartTime    time.Time
	GlobalConfig GlobalConfig
}

// SessionOptions holds session options for initialization
type SessionOptions struct {
	StorageEngine StorageEngine
}

// GlobalConfig holds global settings used by the app
// These may end up just being global constants.
type GlobalConfig struct {
	TTL                 int
	FailedLoginAttempts int
	MaxLoginAttempts    int
}

// NewGlobalConfig creates a new global config struct with default settings.
// This is primarily used for initializing a new data store
func NewGlobalConfig() GlobalConfig {
	return GlobalConfig{
		TTL:                 DefaultSessionTTL,
		FailedLoginAttempts: 0,
		MaxLoginAttempts:    DefaultMaxLoginAttempts,
	}
}

// ToJSON is a helper method for GlobalConfig
func (g GlobalConfig) ToJSON() []byte {
	b, _ := json.Marshal(g)
	return b
}

// NewSession takes a password and opts and returns a pointer to Session and an error
func NewSession(password string, opts SessionOptions) (*Session, error) {
	storageOpts := StorageOptions{Engine: opts.StorageEngine}
	storage, err := NewStorage(storageOpts)
	if err != nil {
		return nil, err
	}

	cipher := NewTimeSeriesSBCipher()
	session := Session{
		Storage:   storage,
		Cipher:    cipher,
		TTL:       DefaultSessionTTL,
		StartTime: time.Now(),
	}

	profilePaths, err := storage.List(ProfileKeyPrefix)
	if err != nil {
		return nil, err
	}
	if len(profilePaths) == 0 {
		return nil, errors.New("no profile found")
	}
	for _, profilePath := range profilePaths {
		id, err := getIDFromPath(profilePath)
		if err != nil {
			return nil, err
		}
		key := DeriveKey([]byte(password), id)
		profile, err := GetProfileFromEncryptedStorage(profilePath, key, cipher, storage)
		if err == nil {
			session.Profile = profile
			return &session, err
		}
	}

	return nil, errors.New("invalid password")
}

// Close gracefully closes the session
func (s *Session) Close() error {
	return s.Storage.Close()
}
