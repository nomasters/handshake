package handshake

import (
	"bytes"
	"encoding/base64"
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
	StorageEngine   StorageEngine
	StorageFilePath string
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
	storageOpts.FilePath = opts.StorageFilePath
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
		if err := s.setLookup(chatID, cp.ID, lookups); err != nil {
			deleteAllWithPrefix(s.storage, basePath)
			return "", err
		}
	}
	if config.PeerID == "" {
		deleteAllWithPrefix(s.storage, basePath)
		return "", errors.New("primary PeerID not found for chat")
	}

	if err := s.setChat(chatID, config); err != nil {
		deleteAllWithPrefix(s.storage, basePath)
		return "", err
	}

	if err := s.setChatlog(chatID, make(chatLog)); err != nil {
		deleteAllWithPrefix(s.storage, basePath)
		return "", err
	}

	s.activeHandshake = &handshake{}
	return chatID, nil
}

// ListChats returns a json encoded list of chatIDs and an error
func (s *Session) ListChats() ([]byte, error) {
	list, err := s.storage.List("chats/")
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(uniqueChatIDsFromPaths(list, s.profile.ID))
}

func (s *Session) getChat(chatID string) (chat, error) {
	key := fmt.Sprintf("chats/%v/%v/config", chatID, s.profile.ID)
	chatGob, err := s.get(key)
	if err != nil {
		return chat{}, err
	}
	return newChatFromGob(chatGob)
}

func (s *Session) setChat(chatID string, c chat) error {
	key := fmt.Sprintf("chats/%v/%v/config", chatID, s.profile.ID)
	safeConfig, err := c.Config()
	if err != nil {
		return err
	}
	chatGob, err := encodeGob(safeConfig)
	if err != nil {
		return err
	}
	_, err = s.set(key, chatGob)
	return err
}

func (s *Session) getLookup(chatID, peerID string) (lookup, error) {
	key := fmt.Sprintf("chats/%v/%v/lookups/%v", chatID, s.profile.ID, peerID)
	lookupGob, err := s.get(key)
	if err != nil {
		return lookup{}, err
	}
	return newLookupFromGob(lookupGob)
}

func (s *Session) setLookup(chatID, peerID string, l lookup) error {
	key := fmt.Sprintf("chats/%v/%v/lookups/%v", chatID, s.profile.ID, peerID)
	lookupGob, err := encodeGob(l)
	if err != nil {
		return err
	}
	_, err = s.set(key, lookupGob)
	return err
}

func (s *Session) GetChatlog(chatID string) (chatLog, error) {
	key := fmt.Sprintf("chats/%v/%v/chatlog", chatID, s.profile.ID)
	chatLogGob, err := s.get(key)
	if err != nil {
		return chatLog{}, err
	}
	return newChatLogFromGob(chatLogGob)
}

func (s *Session) setChatlog(chatID string, cl chatLog) error {
	key := fmt.Sprintf("chats/%v/%v/chatlog", chatID, s.profile.ID)
	chatLogGob, err := encodeGob(cl)
	if err != nil {
		return err
	}
	_, err = s.set(key, chatLogGob)
	return err
}

func (s *Session) getRendezvousHash(chatID, peerID string) (hash string) {
	c, err := s.getChat(chatID)
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := s.getLookup(chatID, peerID)
	if err != nil {
		fmt.Println(err)
		return
	}

	rBytes, err := c.Peers[peerID].Strategy.Rendezvous.Get("")
	if err != nil {
		fmt.Println(err)
		return // TODO: skip for now, there should be more logic here.
	}

	rHash := base64.StdEncoding.EncodeToString(rBytes[:lookupHashLength])
	rKey := l.popKey(rHash)
	if err := s.setLookup(chatID, peerID, l); err != nil {
		fmt.Println(err)
		return
	}
	hashBytes, err := c.Peers[peerID].Strategy.Cipher.Decrypt(rBytes[lookupHashLength:], rKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	hash = string(hashBytes)

	cl, err := s.GetChatlog(chatID)
	if err != nil {
		fmt.Println(err)
		return
	}

	if cl.HashInLog(hash) {
		return ""
	}
	if err := s.setChat(chatID, c); err != nil {
		return ""
	}
	return hash
}

func (s *Session) retrieveMessage(chatID, hash, peerID string) (data chatData, err error) {
	c, err := s.getChat(chatID)
	if err != nil {
		return
	}

	l, err := s.getLookup(chatID, peerID)
	if err != nil {
		fmt.Println(err)
		return
	}

	b, err := c.Peers[peerID].Strategy.Storage.Get(hash)
	if err != nil {
		return
	}
	lookupHash := base64.StdEncoding.EncodeToString(b[:lookupHashLength])
	key := l.popKey(lookupHash)
	if len(key) == 0 {
		return data, errors.New("no key")
	}
	err = s.setLookup(chatID, peerID, l)
	if err != nil {
		return
	}
	d, err := c.Peers[peerID].Strategy.Cipher.Decrypt(b[lookupHashLength:], key)
	if err != nil {
		return
	}
	err = json.Unmarshal(d, &data)
	if err != nil {
		return
	}
	err = s.setChat(chatID, c)
	return
}

func (s *Session) logChatData(chatID string, peerID string, hash string, data chatData) error {
	cl, err := s.GetChatlog(chatID)
	if err != nil {
		return err
	}

	clEntry := chatLogEntry{
		ID:     hash,
		Sender: peerID,
		Sent:   data.Timestamp,
		TTL:    data.TTL,
		Data:   data,
	}

	if err := cl.AddEntry(clEntry); err != nil {
		return err
	}
	b, _ := cl.SortedJSON()
	fmt.Println(string(b))
	return s.setChatlog(chatID, cl)
}

func (s *Session) recursivelyLogParents(chatID string, peerID string, data chatData) error {
	if data.Parent == "" {
		return nil // if no parent set, return early
	}
	cl, err := s.GetChatlog(chatID)
	if err != nil {
		return err
	}
	if cl.HashInLog(data.Parent) {
		return nil // if hash already in log, return early
	}
	parentData, err := s.retrieveMessage(chatID, data.Parent, peerID)
	if err != nil {
		if err.Error() == "no key" {
			return nil
		}
		return err
	}
	if err := s.logChatData(chatID, peerID, data.Parent, parentData); err != nil {
		return err
	}
	if parentData.Parent != "" {
		return s.recursivelyLogParents(chatID, peerID, parentData)
	}
	return nil
}

// RetrieveMessages takes a chatID and initiates the retrieval process for all peers
// it returns a json encoded chatLogList and error
func (s *Session) RetrieveMessages(chatID string) ([]byte, error) {
	// this should query all peer endpoints and update the chatlog
	// this step also runs ttl validation to clear out old messages
	// it returns a json encoded chatLogList

	c, err := s.getChat(chatID)
	if err != nil {
		return []byte{}, err
	}

	for peerID := range c.Peers {
		if peerID == c.PeerID { // skip self
			continue
		}
		hash := s.getRendezvousHash(chatID, peerID)
		if hash == "" {
			fmt.Println("no hash")
			continue
		}
		data, err := s.retrieveMessage(chatID, hash, peerID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if err := s.logChatData(chatID, peerID, hash, data); err != nil {
			fmt.Println(err)
			continue
		}

		if err := s.recursivelyLogParents(chatID, peerID, data); err != nil {
			fmt.Println(err)
			continue
		}

	}
	cl, err := s.GetChatlog(chatID)
	if err != nil {
		return []byte{}, err
	}
	return cl.SortedJSON()
}

// SendMessage takes a chatID and message bytes and submits the message to the message
// storage and rendezvous point. It returns a json encoded chatLogList and error
func (s *Session) SendMessage(chatID string, b []byte) ([]byte, error) {
	if len(b) > maxMessageSize {
		return []byte{}, fmt.Errorf("messag sized exceeds max size of %v bytes", maxMessageSize)
	}

	c, err := s.getChat(chatID)
	if err != nil {
		return []byte{}, err
	}

	var data chatData
	if err := json.Unmarshal(b, &data); err != nil {
		return []byte{}, err
	}
	data.Parent = c.LastSent
	data.Timestamp = time.Now().UnixNano()
	data.TTL = c.TTL()

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return []byte{}, nil
	}

	sender := c.Peers[c.PeerID]

	l, err := s.getLookup(chatID, c.PeerID)
	if err != nil {
		return []byte{}, err
	}
	mStoreKey, mStoreValue := l.popRandom()
	if err := s.setLookup(chatID, c.PeerID, l); err != nil {
		return []byte{}, err
	}

	mStoreKeyBytes, err := base64.StdEncoding.DecodeString(mStoreKey)
	if err != nil {
		return []byte{}, err
	}

	cipherText, err := sender.Strategy.Cipher.Encrypt(dataBytes, mStoreValue)
	if err != nil {
		return []byte{}, err
	}

	var payload []byte
	payload = append(payload, mStoreKeyBytes...)
	payload = append(payload, cipherText...)
	hash, err := sender.Strategy.Storage.Set("", payload)
	if err != nil {
		return []byte{}, err
	}
	c.LastSent = hash
	fmt.Println(c.LastSent)

	if err := s.setChat(chatID, c); err != nil {
		return []byte{}, err
	}

	rStoreKey, rStoreValue := l.popRandom()
	if err := s.setLookup(chatID, c.PeerID, l); err != nil {
		return []byte{}, err
	}

	rStoreKeyBytes, err := base64.StdEncoding.DecodeString(rStoreKey)
	if err != nil {
		return []byte{}, err
	}

	rCipherText, err := sender.Strategy.Cipher.Encrypt([]byte(hash), rStoreValue)
	if err != nil {
		return []byte{}, err
	}

	var rPayload []byte
	rPayload = append(rPayload, rStoreKeyBytes...)
	rPayload = append(rPayload, rCipherText...)

	if _, err := sender.Strategy.Rendezvous.Set("", rPayload); err != nil {
		return []byte{}, err
	}

	cl, err := s.GetChatlog(chatID)
	if err != nil {
		return []byte{}, err
	}

	clEntry := chatLogEntry{
		ID:     hash,
		Sender: c.PeerID,
		Sent:   data.Timestamp,
		TTL:    data.TTL,
		Data:   data,
	}

	if err := cl.AddEntry(clEntry); err != nil {
		return []byte{}, err
	}
	if err := s.setChatlog(chatID, cl); err != nil {
		return []byte{}, err
	}

	return cl.SortedJSON()
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
