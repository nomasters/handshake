package handshake

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nomasters/hashmap"
	bolt "go.etcd.io/bbolt"
)

// StorageEngine type for enum
type StorageEngine int

const (
	// BoltEngine is the default storage engine for device storage
	BoltEngine StorageEngine = iota
	// HashmapEngine is the default Rendezvous storage type
	HashmapEngine
	// IPFSEngine is the default message storage type
	IPFSEngine
)

const (
	// DefaultStorageEngine is used to set the storage engine if none is set in
	// storage options
	defaultStorageEngine = BoltEngine
	// DefaultBoltFilePath is the default path and file name for BoltDB storage
	defaultBoltFilePath = "handshake.boltdb"
	// DefaultTLB is the name of the top level bucket for BoltDB
	defaultTLB = "handshake"
	// GlobalConfigKey is the key string for where global-config is stored
	globalConfigKey      = "global-config"
	maxIPFSRead          = 3000000 // ~3MB
	defaultRendezvousURL = "https://prototype.hashmap.sh"
)

type signatureType int

const (
	// ed25519
	ed25519 signatureType = iota
)

// consensusRule is a datatype to capture basic rules around how consensus with multiple nodes should
// work for storage such as IPFS and Hashmap if multiple endpoints are configured.
type consensusRule int

const (
	// firstSuccess dictates that if any node returns a success, success is returned
	firstSuccess consensusRule = iota
	// redundantPairSuccess dictates that if any two nodes return a success, success is returned
	redundantPairSuccess
	// majoritySuccess dictates that if a simple majority of nodes returns success, a sucess is returned
	majoritySuccess
	// unanimousSuccess dictates that all nodes must return a success to return a sucess
	unanimousSuccess
)

const (
	defaultConsensusRule  = firstSuccess
	defaultHashmapSigType = ed25519
)

// Storage is the primary interface for interacting with the KV store in handshake
type storage interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) (string, error)
	Delete(key string) error
	List(path string) ([]string, error)
	Close() error
	export() (storageConfig, error)
	share() (peerStorage, error)
}

func newDefaultRendezvous() *hashmapStorage {
	privateKey := hashmap.GenerateKey()
	publicKey := privateKey[32:]
	n := node{
		URL: defaultRendezvousURL,
	}
	sig := signatureAlgorithm{
		Type:       ed25519,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
	return &hashmapStorage{
		WriteNodes: []node{n},
		Signatures: []signatureAlgorithm{sig},
		WriteRule:  defaultConsensusRule,
	}
}

func newDefaultMessageStorage() ipfsStorage {
	settings := make(map[string]string)
	settings["query_type"] = "api"

	n := node{
		URL:      "https://ipfs.infura.io:5001/",
		Settings: settings,
	}

	return ipfsStorage{
		WriteNodes: []node{n},
		WriteRule:  defaultConsensusRule,
	}
}

// peerStorage is a set of aggregate settings used for sharing and storing storage settings
type peerStorage struct {
	Type       StorageEngine `json:"type"`
	ReadNodes  []node        `json:"read_nodes,omitempty"`
	WriteNodes []node        `json:"write_nodes,omitempty"`
	ReadRule   consensusRule `json:"read_rule,omitempty"`
	WriteRule  consensusRule `json:"write_rule,omitempty"`
}

// storageConfig is a set of settings used to in storage interface gob storage
type storageConfig struct {
	Type       StorageEngine
	ReadNodes  []node
	WriteNodes []node
	ReadRule   consensusRule
	WriteRule  consensusRule
	Signatures []signatureAlgorithm
	Latest     int64
}

type node struct {
	URL      string            `json:"url,omitempty"`
	Header   map[string]string `json:"header,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
}

// StorageOptions are used to pass in initialization settings
type StorageOptions struct {
	Engine     StorageEngine
	FilePath   string
	Signatures []signatureAlgorithm
	ReadNodes  []node
	WriteNodes []node
	ReadRule   consensusRule
	WriteRule  consensusRule
}

// NewStorage initiates a new storage Interface
func newStorage(opts StorageOptions) (storage, error) {
	switch opts.Engine {
	case BoltEngine:
		return newBoltStorage(opts)
	default:
		return nil, errors.New("invalid engine type")
	}
}

// NewBoltStorage takes StorageOptions as an argument and returns a reference to a BoltDB
// based implementation of the Storage interface.
func newBoltStorage(opts StorageOptions) (boltStorage, error) {
	tlb := defaultTLB
	fp := defaultBoltFilePath
	if opts.FilePath != "" {
		fp = opts.FilePath
	}
	db, err := bolt.Open(fp, 0666, nil)
	if err != nil {
		return boltStorage{}, err
	}

	// ensure that top level bucket exists
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(tlb)); err != nil {
			return fmt.Errorf("error creating bucket: %s", err)
		}
		return nil
	}); err != nil {
		return boltStorage{}, err
	}

	// ensure that globalConfig exists, and if not, initialize GlobalConfig
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tlb))
		blob := b.Get([]byte(globalConfigKey))
		if blob == nil {
			return b.Put([]byte(globalConfigKey), newGlobalConfig().ToJSON())
		}
		return nil
	}); err != nil {
		return boltStorage{}, err
	}

	return boltStorage{db: db, tlb: tlb}, nil
}

// BoltStorage is a struct that conforms to the Storage interface for using
// BoltDB. DB is a reference to a boltDB instance and TLB stands for "top level bucket"
type boltStorage struct {
	db  *bolt.DB
	tlb string
}

// Get takes a key string and returns a byte slice or error from a BoltStorage struct. Get
// get retruns and empty byte slice if no key is found and/or the byte slice is blank. An error
// is returned if the key invalid in formatting, it is too long, or there is an underlying issue
// with boltDB
func (s boltStorage) Get(key string) (value []byte, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.tlb))
		value = b.Get([]byte(key))
		return nil
	})
	return value, err
}

// Set takes a key string and value byte slice returns an error from a BoltStorage struct.
// Set treats both create and updates the same. Errors are returned if the key has invalid syntax
// and if key or value are too long.
func (s boltStorage) Set(key string, value []byte) (string, error) {
	return key, s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.tlb))
		return b.Put([]byte(key), value)
	})
}

// Delete takes a key string and deletes item, if it exists in storage, returns an error from a BoltStorage struct.
func (s boltStorage) Delete(key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.tlb))
		return b.Delete([]byte(key))
	})
}

// List takes a path and returns a slice of key paths formatted as strings or an error.
func (s boltStorage) List(path string) (keys []string, err error) {
	p := []byte(path)
	err = s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(s.tlb)).Cursor()
		for k, _ := c.Seek(p); k != nil && bytes.HasPrefix(k, p); k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		return nil
	})
	return keys, err
}

// share is not configured on BoltStorage, since it is private storage.
// Therefore it returns an empty struct.
func (s boltStorage) share() (peerStorage, error) {
	return peerStorage{}, errors.New("this storage does not support shared configs")
}

// share is not configured on BoltStorage, since it is private storage.
// Therefore it returns an empty struct.
func (s boltStorage) export() (storageConfig, error) {
	return storageConfig{}, errors.New("this storage does not support exporting configs")
}

// Close is used to close the Bolt DB engine and returns an error
func (s boltStorage) Close() error {
	return s.db.Close()
}

// HashmapStorage interacts with a hashmap server and
// conforms to the Storage interface
type hashmapStorage struct {
	ReadNodes  []node
	WriteNodes []node
	Signatures []signatureAlgorithm
	ReadRule   consensusRule
	WriteRule  consensusRule
	Latest     int64
}

type signatureAlgorithm struct {
	Type       signatureType
	PrivateKey []byte
	PublicKey  []byte
}

func newHashmapStorage(opts StorageOptions) (*hashmapStorage, error) {
	return &hashmapStorage{
		Signatures: opts.Signatures,
		ReadNodes:  opts.ReadNodes,
		WriteNodes: opts.WriteNodes,
		ReadRule:   opts.ReadRule,
		WriteRule:  opts.WriteRule,
	}, nil
}

func (s *hashmapStorage) updateLatest(timeStamp int64) error {
	// check for timestamp set too far in the future
	if timeStamp > (time.Now().UnixNano() + (5 * 1000000000)) {
		return errors.New("invalid future timestamp")
	}
	// check for potential replay attack, which latest timestamp
	// detected newer than the one provided by the server
	if s.Latest > timeStamp {
		return errors.New("stale timestamp")
	}
	s.Latest = timeStamp
	return nil
}

// getHashFromPath takes a path string and returns the hash at the end of the path
func getHashFromPath(path string) string {
	lastIndex := strings.LastIndex(path, "/")
	if lastIndex == -1 {
		return path
	}
	return path[lastIndex+1:]
}

// getFirstSuccess loops through all ReadNodes in a hashmapStorage and attempts to resolve the data from a
// payload. There is an important set of steps that this goes through, including:
// - validating the MultiHash in the URL is supported
// - comparing the payload pubkey to the url hash, which must match.
// if all verification and validations are successful, it returns the data bytes from the payload
func (s *hashmapStorage) getFirstSuccess() ([]byte, error) {
	for _, node := range s.ReadNodes {
		u, err := url.Parse(node.URL)
		if err != nil {
			return []byte{}, fmt.Errorf("invalid url for: %v", node.URL)
		}
		urlHash := getHashFromPath(u.Path)
		if !isHashmapMultihash(urlHash) {
			return []byte{}, fmt.Errorf("invalid hashmap endpoint for: %v", node.URL)
		}

		resp, err := http.Get(node.URL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		payload, err := hashmap.NewPayloadFromReader(resp.Body)
		if err != nil {
			continue
		}

		pubkey, err := payload.PubKeyBytes()
		if err != nil {
			return []byte{}, fmt.Errorf("invalid pubkey in payload for: %v", node.URL)
		}

		if urlHash != base58Multihash(pubkey) {
			return []byte{}, fmt.Errorf("payload and endpoint hash mismatch for: %v", node.URL)
		}

		data, err := payload.GetData()
		if err != nil {
			return []byte{}, err
		}
		if err := s.updateLatest(data.Timestamp); err != nil {
			return []byte{}, err
		}
		return data.MessageBytes()
	}
	return []byte{}, errors.New("no servers available")
}

func (s *hashmapStorage) Get(key string) ([]byte, error) {
	if len(s.ReadNodes) < 1 {
		return []byte{}, errors.New("no read nodes configured")
	}
	switch s.ReadRule {
	case firstSuccess:
		return s.getFirstSuccess()
	default:
		return []byte{}, errors.New("This readRule is not yet implemented")
	}

}

func (s *hashmapStorage) setFirstSuccess(payload []byte) error {
	for _, node := range s.WriteNodes {
		resp, err := http.Post(node.URL, "application/json", bytes.NewReader(payload))
		if err != nil {
			continue
		}
		if resp.StatusCode > 399 {
			continue
		}
		return nil
	}
	return errors.New("no servers available")
}

func (s *hashmapStorage) Set(key string, value []byte) (string, error) {
	if len(s.WriteNodes) < 1 {
		return key, errors.New("no write nodes configured")
	}

	opts := hashmap.GeneratePayloadOptions{Message: string(value)}
	// TODO: currently we only support one signature, but this will change
	payload, err := hashmap.GeneratePayload(opts, s.Signatures[0].PrivateKey)
	if err != nil {
		return key, err
	}

	switch s.WriteRule {
	case firstSuccess:
		return key, s.setFirstSuccess(payload)
	default:
		return key, errors.New("This writeRule is not yet implemented")
	}
}

// Delete is used to remove references from hashmap. Not currently implemented.
// TODO : a delete could be accomplished by writing a blank dataset to each endpoint
func (s hashmapStorage) Delete(key string) (e error) { return }

// List is not implimented for hashmapStorage, returns "", nil
func (s hashmapStorage) List(path string) ([]string, error) {
	return []string{}, errors.New("no implimented")
}

// Close is not used in hashmap, returns nil
func (s hashmapStorage) Close() (e error) { return }

// share returns a peerStorage and error, it generates read nodes from the write nodes + pubkey
// it also returns ReadRules based on the WriteRules
func (s hashmapStorage) share() (peerStorage, error) {
	readNodes, err := s.genReadFromWriteNodes()
	if err != nil {
		return peerStorage{}, err
	}

	return peerStorage{
		Type:      HashmapEngine,
		ReadNodes: readNodes,
		ReadRule:  s.WriteRule,
	}, nil
}

// TODO: configure export settings for this
func (s hashmapStorage) export() (storageConfig, error) {
	return storageConfig{
		Type:       HashmapEngine,
		ReadNodes:  s.ReadNodes,
		WriteNodes: s.WriteNodes,
		ReadRule:   s.ReadRule,
		WriteRule:  s.WriteRule,
		Signatures: s.Signatures,
		Latest:     s.Latest,
	}, nil
}

// genReadFromWriteNodes creates a set of read nodes based on all signature
// files times the number of write urls and returns a list of nodes and and error
func (s hashmapStorage) genReadFromWriteNodes() ([]node, error) {
	var readNodes []node
	var endpoints []string
	for _, sig := range s.Signatures {
		endpoints = append(endpoints, base58Multihash(sig.PublicKey))
	}
	for _, writeNode := range s.WriteNodes {
		for _, endpoint := range endpoints {
			u, err := url.Parse(writeNode.URL)
			if err != nil {
				return readNodes, err
			}
			u.Path = endpoint
			readNodes = append(readNodes, node{URL: u.String()})
		}
	}
	return readNodes, nil
}

// IPFSStorage interacts with an IPFS gateway and conforms to the
// Storage interface
type ipfsStorage struct {
	ReadNodes  []node
	WriteNodes []node
	ReadRule   consensusRule
	WriteRule  consensusRule
}

func newIPFSStorage(opts StorageOptions) (ipfsStorage, error) {
	return ipfsStorage{
		ReadNodes:  opts.ReadNodes,
		WriteNodes: opts.WriteNodes,
		ReadRule:   opts.ReadRule,
		WriteRule:  opts.WriteRule,
	}, nil
}

func (s ipfsStorage) Get(key string) ([]byte, error) {
	if len(s.ReadNodes) < 1 {
		return []byte{}, errors.New("no read nodes configured")
	}
	switch s.ReadRule {
	case firstSuccess:
		return s.getFirstSuccess(key)
	default:
		return []byte{}, errors.New("This readRule is not yet implemented")
	}
}

func (s *ipfsStorage) getFirstSuccess(hash string) ([]byte, error) {
	for _, node := range s.ReadNodes {
		resp, err := getFromIPFS(node, hash)
		if err != nil {
			continue
		}
		return resp, nil
	}
	return []byte{}, errors.New("no servers available")
}

func (s ipfsStorage) Set(key string, value []byte) (string, error) {
	if len(s.WriteNodes) < 1 {
		return "", errors.New("no write nodes configured")
	}
	switch s.WriteRule {
	case firstSuccess:
		return s.setFirstSuccess(value)
	default:
		return "", errors.New("This writeRule is not yet implemented")
	}
}

func (s ipfsStorage) setFirstSuccess(body []byte) (string, error) {
	for _, node := range s.WriteNodes {
		resp, err := postToIPFS(node, body)
		if err != nil {
			continue
		}
		return resp, nil
	}
	return "", errors.New("no servers available")
}

func (s ipfsStorage) Delete(key string) error            { return nil }
func (s ipfsStorage) List(path string) ([]string, error) { return []string{}, nil }
func (s ipfsStorage) Close() error                       { return nil }

func (s ipfsStorage) share() (peerStorage, error) {
	return peerStorage{
		Type:      IPFSEngine,
		ReadNodes: s.WriteNodes,
		ReadRule:  s.WriteRule,
	}, nil
}

// TODO: configure export settings for this
func (s ipfsStorage) export() (storageConfig, error) {
	return storageConfig{
		Type:       IPFSEngine,
		ReadNodes:  s.ReadNodes,
		ReadRule:   s.ReadRule,
		WriteNodes: s.WriteNodes,
		WriteRule:  s.WriteRule,
	}, nil
}

func newStorageFromPeer(s peerStorage) (storage, error) {
	switch s.Type {
	case IPFSEngine:
		return ipfsStorage{
			ReadNodes: s.ReadNodes,
			ReadRule:  s.ReadRule,
		}, nil
	case HashmapEngine:
		return &hashmapStorage{
			ReadNodes: s.ReadNodes,
			ReadRule:  s.ReadRule,
		}, nil
	default:
		return nil, errors.New("invalid storage engine type")
	}
}

func newStorageFromConfig(s storageConfig) (storage, error) {
	switch s.Type {
	case IPFSEngine:
		return ipfsStorage{
			ReadNodes:  s.ReadNodes,
			ReadRule:   s.ReadRule,
			WriteNodes: s.WriteNodes,
			WriteRule:  s.WriteRule,
		}, nil
	case HashmapEngine:
		return &hashmapStorage{
			ReadNodes:  s.ReadNodes,
			ReadRule:   s.ReadRule,
			WriteNodes: s.WriteNodes,
			WriteRule:  s.WriteRule,
			Signatures: s.Signatures,
			Latest:     s.Latest,
		}, nil
	default:
		return nil, errors.New("invalid storage engine type")
	}
}

// TODO: these should prob be moved into their own lib.

// appendToPath this safely appends two url paths together by ensuring that leading and trailing
// slashes are trimmed before joining them together
func appendToPath(base, add string) string {
	if add == "" {
		return base
	}
	base = strings.TrimSuffix(base, "/")
	add = strings.TrimPrefix(add, "/")
	return fmt.Sprintf("%s/%s", base, add)
}

func getFromIPFS(n node, hash string) ([]byte, error) {
	client := http.DefaultClient
	u, err := url.Parse(n.URL)
	if err != nil {
		return []byte{}, err
	}
	switch n.Settings["query_type"] {
	case "api":
		endpoint := "api/v0/cat"
		values := u.Query()
		values.Set("arg", hash)
		u.RawQuery = values.Encode()
		u.Path = appendToPath(u.Path, endpoint)
	default:
		endpoint := fmt.Sprintf("ipfs/%s", hash)
		u.Path = appendToPath(u.Path, endpoint)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return []byte{}, err
	}
	if len(n.Header) > 0 {
		for k, v := range n.Header {
			req.Header.Set(k, v)
		}
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	limitedReader := &io.LimitedReader{R: resp.Body, N: maxIPFSRead}
	return ioutil.ReadAll(limitedReader)
}

func postToIPFS(n node, body []byte) (string, error) {
	client := http.DefaultClient
	u, err := url.Parse(n.URL)
	if err != nil {
		return "", err
	}
	switch n.Settings["query_type"] {
	case "api":
		endpoint := "api/v0/add"
		u.Path = appendToPath(u.Path, endpoint)
		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)
		fileWriter, err := bodyWriter.CreateFormFile("file", "file")
		if err != nil {
			return "", err
		}
		if _, err := fileWriter.Write(body); err != nil {
			return "", err
		}
		contentType := bodyWriter.FormDataContentType()
		bodyWriter.Close()
		req, err := http.NewRequest("POST", u.String(), bodyBuf)
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", contentType)
		if len(n.Header) > 0 {
			for k, v := range n.Header {
				req.Header.Set(k, v)
			}
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		output := make(map[string]string)
		if err := json.Unmarshal(body, &output); err != nil {
			return "", err
		}
		return output["Hash"], nil
	default:
		endpoint := "ipfs/"
		u.Path = appendToPath(u.Path, endpoint)
		req, err := http.NewRequest("POST", u.String(), bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		if len(n.Header) > 0 {
			for k, v := range n.Header {
				req.Header.Set(k, v)
			}
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		return resp.Header.Get("Ipfs-Hash"), nil
	}
}
