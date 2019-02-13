/*
Package handshake is to describe the basic interfaces,
methods, and funcationality for handshake core. This is subject
to change and will change greatly as ideas solidify
*/
package handshakespec

// ideas for structs

// Session is the primary struct for a logged in  user. It holds the profile data
// as well as settings information
type Session struct {
	Profile      Profile
	Storage      *Storage
	TTL          int64
	GlobalConfig GlobalConfig
}

// StorageEngine type for enum
type StorageEngine int

type CipherType int

const (
	// BoltEngine is the first supported StorageEngine
	BoltEngine StorageEngine = iota
)

const (
	SecretBox CipherType = iota
	oneTimePad
)

type ProfileType int

const (
	// PrimaryProfile is the profile type used by a logged in user
	PrimaryProfile ProfileType = iota
	// DuressProfile is the profile type used by dummy-duress accounts
	DuressProfile
)

const (
	// DefaultStorageEngine is the default engine used if no override is set
	DefaultStorageEngine = BoltEngine
	// DefaultSessionTTL is the default TTL before a Session closes
	DefaultSessionTTL = 300 // 5 minutes in seconds
	// DefaultMaxLoginAttempts is the number of times failed login attempts are allowed
	DefaultMaxLoginAttempts = 10
)

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

// NewSession takes a password and opts and returns a pointer to Session
func NewSession(password string, opts SessionOptions) *Session {}

// Close gracefully closes the session
func (s *Session) Close() {}

// Profile represents a profile that has been accessed
// this would contain successfully decrypted profile data
type Profile struct {
	ID        string
	Type      ProfileType
	key       string
	Delegated []DelegatedProfile
	Settings  ProfileSettings
}

// ProfileSettings holds profile settings info
type ProfileSettings struct {
	SessionTTL int64
}

// DelegatedProfile holds delegated profile information
type DelegatedProfile struct {
	ID   string
	Type string
	Key  string
}

// ChatSession
type ChatSession struct {
	ID            string
	Config        ChatConfig
	ParticipantID string
	Participants  map[string]ChatParticipant
	Storage       *Storage
}

type ChatConfig struct {
	MaxTTL int
}

type ChatParticipant struct {
	ID       string
	Alias    string
	Strategy Strategy
}

type Strategy struct {
	Rendevous MessageRendevous
	Storage   MessageStorage
	Cipher    MessageCipher
}

type MessageRendevous interface {
	GetRendevous() ([]byte, error)
	PostRendevous(message []byte) ([]byte, error)
}

type MessageStorage interface {
	GetMessage(key string) ([]byte, error)
	PostMessage(message []byte) ([]byte, error)
}

type MessageCipher interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type HashMapRendevous struct {
	Type       string
	Nodes      []string
	SigMethods []HashmapSigMethod
	// Possibly an consensus method? any or all
	ConsensusRule  string
	PublishingRule string
}

type HashmapSigMethod struct {
	Type            string
	PrivateKeys     [][]byte
	MultiHashPubkey [][]byte
}

type IPFSMessageStorage struct {
	Type           string
	ReadNodes      []string
	WriteNodes     []string
	ConsensusRule  string
	PublishingRule string
}

type SecretBoxCipher struct {
	Type      string
	ChunkSize int
	Lookups   string
	Storage   *Storage
}

/*


 */

// type RendevousStrategy struct {
// 	Type string
// }

type ChatLog struct {
	ID       string
	Sender   string
	Sent     int64
	Received int64
	TTL      int64
	Data     []byte
}

// Lookup is a map of string string, it holds a lookupKey and a one-time key
// in string format
type Lookup map[string]string

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
}

// NewStorage initiates a new storage Interface
func NewStorage(opts StorageOptions) (*Storage, error) {}
