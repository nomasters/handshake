package handshake

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// DefaultSessionTTL is the default TTL before a Session closes
	DefaultSessionTTL = 15 * 60 // 15 minutes in seconds
	// DefaultMaxLoginAttempts is the number of times failed login attempts are allowed
	DefaultMaxLoginAttempts = 10
	chatIDLength            = 12
	defaultLookupCount      = 10000
)

// Session is the primary struct for a logged in  user. It holds the profile data
// as well as settings information
type Session struct {
	profile         Profile
	storage         storage
	cipher          cipher
	ttl             int64
	startTime       int64
	globalConfig    globalConfig
	activeHandshake *handshake
}

// SessionOptions holds session options for initialization
type SessionOptions struct {
	StorageEngine StorageEngine
}

// GlobalConfig holds global settings used by the app
// These may end up just being global constants.
type globalConfig struct {
	TTL                 int
	FailedLoginAttempts int
	MaxLoginAttempts    int
}

// newGlobalConfig creates a new global config struct with default settings.
// This is primarily used for initializing a new data store
func newGlobalConfig() globalConfig {
	return globalConfig{
		TTL:                 DefaultSessionTTL,
		FailedLoginAttempts: 0,
		MaxLoginAttempts:    DefaultMaxLoginAttempts,
	}
}

// ToJSON is a helper method for GlobalConfig
func (g globalConfig) ToJSON() []byte {
	b, _ := json.Marshal(g)
	return b
}

// NewSession takes a password and opts and returns a pointer to Session and an error
func NewSession(password string, opts SessionOptions) (*Session, error) {
	storageOpts := StorageOptions{Engine: opts.StorageEngine}
	storage, err := newStorage(storageOpts)
	if err != nil {
		return nil, err
	}

	cipher := newTimeSeriesSBCipher()
	session := Session{
		storage:   storage,
		cipher:    cipher,
		ttl:       DefaultSessionTTL,
		startTime: time.Now().Unix(),
	}

	profilePaths, err := storage.List(profileKeyPrefix)
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
		key := deriveKey([]byte(password), id)
		profile, err := getProfileFromEncryptedStorage(profilePath, key, cipher, storage)
		if err == nil {
			session.setProfile(profile)
			return &session, err
		}
	}

	return nil, errors.New("invalid password")
}

// NewDefaultSession is a wrapper around NewSession and applies simple defaults. This is intended to be used
//by the reference apps.
func NewDefaultSession(password string) (*Session, error) {
	opts := SessionOptions{StorageEngine: defaultStorageEngine}
	return NewSession(password, opts)
}

// setProfile takes a profile and sets it to the private variable in the Session struct
func (s *Session) setProfile(p Profile) {
	s.profile = p
}

// GetProfile returns the profile in the Session struct
func (s *Session) GetProfile() Profile {
	return s.profile
}

// Close gracefully closes the session
func (s *Session) Close() error {
	return s.storage.Close()
}

// NewInitiatorWithDefaults provides a simple method with no arguments to create a default handshake
// for an initiator. Adds this handshake pointer to the ActiveHandshake in the session.
func (s *Session) NewInitiatorWithDefaults() {
	s.activeHandshake = newHandshakeInitiatorWithDefaults()
}

// NewPeerWithDefaults provides a simple method with no arguments to create a default handshake
// for an peer. Adds this handshake pointer to the ActiveHandshake in the session.
func (s *Session) NewPeerWithDefaults() {
	s.activeHandshake = newHandshakePeerWithDefaults()
}

// ShareHandshakePosition returns the values from negotiator.Share() from the ActiveHandshake
func (s *Session) ShareHandshakePosition() (b []byte, err error) {
	// TODO: add encryption wrapper
	return s.activeHandshake.Position.Share()
}

// AddPeerToHandshake takes a json encoded peerConfig, attempts to unmarshal it and add it as a peer.
// It returns a bool and an error. The bool indicates if handshake.AllPeersReceived == true, in which case
// the handshake can safely be conversted int a chat.
func (s *Session) AddPeerToHandshake(body []byte) (bool, error) {
	// TODO: add decryption wrapper
	var config peerConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return false, err
	}
	if err := s.activeHandshake.AddPeer(config); err != nil {
		return false, err
	}
	return s.activeHandshake.AllPeersReceived(), nil
}

// GetHandshakePeerTotal returns an int count of the number of peers to expect for a handshake
func (s *Session) GetHandshakePeerTotal() int {
	return s.activeHandshake.GetPeerTotal()
}

// GetHandshakePeerConfig returns the json bytes encoded peerConfig based on peerID or and an error
func (s *Session) GetHandshakePeerConfig(sortNumber int) ([]byte, error) {
	configs, err := s.activeHandshake.GetAllConfigs()
	if err != nil {
		return []byte{}, err
	}
	if sortNumber <= 0 {
		return []byte{}, errors.New("sortNumber must be greater than 0")
	}
	if sortNumber > len(configs) {
		return []byte{}, errors.New("sortNumber is out of range")
	}
	return json.Marshal(configs[sortNumber-1])
}

// set is a wrapper for combining the cipher and storage interfaces. Data in the value component is encrypted and then
// stored in the storage engine.
func (s *Session) set(key string, value []byte) (string, error) {
	encrypted, err := s.cipher.Encrypt(value, s.profile.Key)
	if err != nil {
		return "", err
	}
	return s.storage.Set(key, encrypted)
}

// get is a wrapper for combining the cipher and storage interfaces. Retrieved data is decrypted and returned
// unencrypted as a byte slice and error
func (s *Session) get(key string) ([]byte, error) {
	encrypted, err := s.storage.Get(key)
	if err != nil {
		return []byte{}, err
	}
	return s.cipher.Decrypt(encrypted, s.profile.Key)
}

// NewChat creates a new chat from the activeHandshake and returns a chat ID string and error.
// If the chat is successfully created, it deletes the contents of the activeHandshake
func (s *Session) NewChat() (string, error) {
	peerTotal := s.GetHandshakePeerTotal()
	negotiatorCount := len(s.activeHandshake.Negotiators)
	if peerTotal < 2 {
		return "", errors.New("not enough peers to start a chat")
	}
	if peerTotal != negotiatorCount {
		return "", fmt.Errorf("expected peer total to be %v but counted %v", peerTotal, negotiatorCount)
	}
	chatID := hex.EncodeToString(genRandBytes(chatIDLength))
	negotiators, err := s.activeHandshake.SortedNegotiatorList()
	if err != nil {
		return "", err
	}
	pepper := generatePepper(negotiators)
	config := chat{
		ID:    chatID,
		Peers: make(map[string]chatPeer),
	}
	basePath := fmt.Sprintf("chats/%v/%v", chatID, s.profile.ID)
	for _, n := range negotiators {
		cp := chatPeer{
			ID:       hex.EncodeToString(genRandBytes(chatIDLength)),
			Alias:    n.Alias,
			Strategy: n.Strategy,
		}
		config.Peers[cp.ID] = cp
		if bytes.Equal(n.Entropy, s.activeHandshake.Position.Entropy) {
			config.PeerID = cp.ID
		}
		var p [64]byte
		var e [96]byte
		copy(p[:], pepper)
		copy(e[:], n.Entropy)
		// TODO support cipherType inspection
		lookups, err := genLookups(p, e, SecretBox, defaultLookupCount)
		if err != nil {
			return "", err
		}
		lookupsPath := fmt.Sprintf("%v/lookups/%v", basePath, cp.ID)
		encodedLookup, err := encodeGob(lookups)
		if err != nil {
			return "", err
		}
		if _, err := s.set(lookupsPath, encodedLookup); err != nil {
			deleteAllWithPrefix(s.storage, basePath)
			return "", err
		}
	}
	if config.PeerID == "" {
		deleteAllWithPrefix(s.storage, basePath)
		return "", errors.New("primary PeerID not found for chat")
	}
	safeConfig, err := config.Config()
	if err != nil {
		deleteAllWithPrefix(s.storage, basePath)
		return "", err
	}
	encodedConfig, err := encodeGob(safeConfig)
	if err != nil {
		deleteAllWithPrefix(s.storage, basePath)
		return "", err
	}
	configPath := fmt.Sprintf("%v/config", basePath)
	if _, err := s.set(configPath, encodedConfig); err != nil {
		deleteAllWithPrefix(s.storage, basePath)
		return "", err
	}
	s.activeHandshake = &handshake{}
	return chatID, nil
}

// deleteAllWithPrefix takes a storage interface and a prefix string. It looks up all keys that
// match the prefix and attempts to run the Delete method on all keys, returns a error or nil.
func deleteAllWithPrefix(s storage, prefix string) error {
	keys, err := s.List(prefix)
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err := s.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

// gobBytes takes an empty interface and returns a byte slice and error
func encodeGob(x interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(x)
	return buffer.Bytes(), err
}

func newLookupFromGob(b []byte) (lookup, error) {
	l := make(lookup)
	var buffer bytes.Buffer
	buffer.Write(b)
	err := gob.NewDecoder(&buffer).Decode(&l)
	return l, err
}

type lookup map[string][]byte

type chat struct {
	ID       string
	PeerID   string
	Peers    map[string]chatPeer
	Settings chatSettings
}

// a chatConfig allows safe encoding of a chat
type chatConfig struct {
	ID       string
	PeerID   string
	Peers    map[string]chatPeerConfig
	Settings chatSettings
}

func (c chat) Config() (chatConfig, error) {
	config := chatConfig{
		ID:       c.ID,
		PeerID:   c.PeerID,
		Peers:    make(map[string]chatPeerConfig),
		Settings: c.Settings,
	}

	for _, peer := range c.Peers {
		peerConfig, err := peer.Config()
		if err != nil {
			return chatConfig{}, err
		}
		config.Peers[peer.ID] = peerConfig
	}
	return config, nil
}

// func (c chat) encodeGob() []byte {

// }

type chatSettings struct {
	MaxTTL int
}

type chatPeer struct {
	ID       string
	Alias    string
	Strategy strategy
}

// Config returns a storage-safe chatPeerConfig and an error
func (c chatPeer) Config() (chatPeerConfig, error) {
	config := chatPeerConfig{
		ID:    c.ID,
		Alias: c.Alias,
	}
	s, err := c.Strategy.Export()
	config.Strategy = s
	return config, err
}

type chatPeerConfig struct {
	ID       string
	Alias    string
	Strategy strategyConfig
}

// entropy_1 = [96 bytes of random data]
// entropy_2 = [96 bytes of random data]

// pepper = blake2b-512(entropy_1[:32]|entropy_2[:32])
// lookup_hashes_1 = argon2(password=pepper, salt=entropy_1[32:64])
// lookup_hashes_2 = argon2(password=pepper, salt=entropy_2[32:64])
// key = argon2(password=entropy_1[:32], salt=entropy_1[64:])

// chats/{chat_id}/{profile_id}/config  <- holds strategy info, etc
// chats/{chat_id}/{profile_id}/chatlog <- holds chat log data
// chats/{chat_id}/{profile_id}/lookups/{peer_id} <- holds json file (or gob) of lookup keys

// TODO:
// - generate chat ID
// -

// New CHAT
// Generate Chat key
// - This should hold the handshake until the handshake is complete
// - This should
