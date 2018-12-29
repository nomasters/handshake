package handshake

import "errors"

const (
	// DefaultSessionTTL is the default TTL before a Session closes
	DefaultSessionTTL = 300 // 5 minutes in seconds
	// DefaultMaxLoginAttempts is the number of times failed login attempts are allowed
	DefaultMaxLoginAttempts = 10
)

// Session is the primary struct for a logged in  user. It holds the profile data
// as well as settings information
type Session struct {
	Profile      Profile
	Key          []byte
	Storage      Storage
	Cipher       Cipher
	TTL          int64
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

// NewSession takes a password and opts and returns a pointer to Session and an error
func NewSession(password string, opts SessionOptions) (*Session, error) {
	storageOpts := StorageOptions{Engine: opts.StorageEngine}
	storage, err := NewStorage(storageOpts)
	if err != nil {
		return nil, err
	}

	cipher := NewTimeSeriesSBCipher()
	session := Session{
		Storage: storage,
		Cipher:  cipher,
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
			session.Key = key
			break
		}
	}

	if len(session.Key) == 0 {
		return nil, errors.New("invalid password")
	}

	return &session, err
}

// Close gracefully closes the session
func (s *Session) Close() error {
	return s.Storage.Close()
}
