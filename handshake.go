package handshake

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"golang.org/x/crypto/blake2b"
)

type role int

const (
	defaultEntropyBytes = 96
	// Version is the hard coded version of handshake-core running
	Version = "0.0.1"
)

const (
	initiator role = iota
	peer
)

type handshake struct {
	Role        role
	Negotiators []negotiator
	Config      handshakeConfig
	Position    negotiator
	PeerTotal   int
}

type handshakeConfig struct {
	Version string
}

type negotiator struct {
	Entropy   []byte
	Alias     string
	Strategy  strategy
	SortOrder int
}

type peerConfig struct {
	Entropy    string         `json:"entropy"`
	Alias      string         `json:"alias"`
	Config     strategyConfig `json:"config"`
	Item       int            `json:"item,omitempty"`
	TotalItems int            `json:"total_items,omitempty"`
}

// AddPeer takes a peerConfig and adds it to a handshake negotiator slice. It checks for unique Entropy bytes.
func (h *handshake) AddPeer(config peerConfig) error {
	if h.Role == peer {
		if config.Item == 0 {
			return errors.New("missing sort order")
		}
		if config.Item > config.TotalItems {
			return errors.New("sort oder id is greater than total size")
		}
		h.PeerTotal = config.TotalItems
	}

	n, err := newNegotiatorFromConfig(config)
	if err != nil {
		return err
	}
	// ensure that the same peer isn't added twice
	for _, negotiator := range h.Negotiators {
		if bytes.Equal(negotiator.Entropy, n.Entropy) {
			return errors.New("duplicate detected, peer must be unique")
		}
	}
	h.Negotiators = append(h.Negotiators, n)

	if h.Role == peer {
		if config.Item == 1 && config.TotalItems == 2 {
			h.Negotiators = append(h.Negotiators, h.Position)
			h.Negotiators[1].SortOrder = 2
		}
	}

	if h.Role == initiator {
		h.PeerTotal = len(h.Negotiators)
	}
	return nil
}

// Share returns JSON encoded bytes of a peerConfig and an error
func (n negotiator) Share() (b []byte, err error) {
	config, err := n.PeerConfig()
	if err != nil {
		return
	}
	b, err = json.Marshal(config)
	if err != nil {
		return
	}
	return
}

func (n negotiator) PeerConfig() (config peerConfig, err error) {
	stratConfig, err := n.Strategy.Config()
	if err != nil {
		return
	}
	config = peerConfig{
		Entropy: base64.StdEncoding.EncodeToString(n.Entropy),
		Alias:   n.Alias,
		Config:  stratConfig,
	}
	return
}

// AllPeersReceived checks the handshake state to see if PeerTotal and Negotiator counts match.
// True is only returned in the case that len(h.Negotiators) and h.PeerTotal are greater than 0 and their ints match.
func (h *handshake) AllPeersReceived() bool {
	nCount := len(h.Negotiators)
	if nCount == 0 {
		return false
	}
	if h.PeerTotal == 0 {
		return false
	}
	if nCount == h.PeerTotal {
		return true
	}
	return false
}

// GetAllConfigs returns an array of all configs needed for a handshake and includes item counts.
func (h *handshake) GetAllConfigs() (configs []peerConfig, err error) {
	// only the initiator should run this function
	if h.Role != initiator {
		return configs, errors.New("only an initiator can GetAllConfigs")
	}
	totalItems := len(h.Negotiators)
	// check for at least two peers
	if totalItems <= 1 {
		return configs, errors.New("at least two peers must be present to ShareAll")
	}
	// check that Position and the first negotiator entropy bytes match
	if !bytes.Equal(h.Position.Entropy, h.Negotiators[0].Entropy) {
		return configs, errors.New("initiator sort order mismatch")
	}
	h.PeerTotal = totalItems
	for i, n := range h.Negotiators {
		itemNumber := i + 1
		c, err := n.PeerConfig()
		if err != nil {
			return configs, err
		}
		c.Item = itemNumber
		c.TotalItems = totalItems
		configs = append(configs, c)
		h.Negotiators[i].SortOrder = itemNumber
	}
	return configs, nil
}

func (h *handshake) SortedNegotiatorList() ([]negotiator, error) {
	totalItems := len(h.Negotiators)
	if totalItems <= 1 {
		return []negotiator{}, errors.New("at least two peers must be present to sort the list")
	}
	negotiators := make([]negotiator, totalItems)
	if totalItems != h.PeerTotal {
		return []negotiator{}, errors.New("peerTotal and slice total mismatch")
	}
	for i, n := range h.Negotiators {
		if (n.SortOrder < 1) || (n.SortOrder > totalItems) {
			return []negotiator{}, errors.New("invalid sort order item")
		}
		negotiators[n.SortOrder] = n
	}
	for i, n := range negotiators {
		if n.SortOrder != i+1 {
			return []negotiator{}, errors.New("invalid sort validation")
		}
	}
	return negotiators, nil
}

// generatePepper takes a list of negotiators and generates a blake2b-512 hash form the first 32 bytes of each entry set.
func generatePepper(negotiators []negotiator) []byte {
	var b []byte
	for _, n := range negotiators {
		b = append(b, n.Entropy[:32]...)
	}
	h := blake2b.Sum512(b)
	return h[:]
}

// GetPeerTotal returns the total count of peers expected for handshake exchange.
// If no peers are present, it returns 0, meaning peer count is invalid
// If 1 peer is present, it returns 1, for simplified exchange between two parties
// If 2 or more peers are present, it returns n + 1 as the total, to deal with sort order
func (h *handshake) GetPeerTotal() int {
	if len(h.Negotiators) <= 1 {
		return len(h.Negotiators)
	}
	return len(h.Negotiators) + 1
}

type handshakeOptions struct {
	Role  role
	Alias string
}

func genPosition() negotiator {
	return negotiator{
		Entropy: genRandBytes(defaultEntropyBytes),
		Alias:   genAlias(),
	}
}

func newHandshake(strategy strategy, opts handshakeOptions) *handshake {
	position := genPosition()
	position.Strategy = strategy
	if opts.Alias != "" {
		position.Alias = opts.Alias
	}

	h := handshake{
		Role:     opts.Role,
		Position: position,
		Config: handshakeConfig{
			Version: Version,
		},
	}
	if h.Role == initiator {
		h.Negotiators = append(h.Negotiators, position)
	}
	return &h
}

func newHandshakeInitiatorWithDefaults() *handshake {
	opts := handshakeOptions{
		Role: initiator,
	}
	return newHandshake(newDefaultStrategy(), opts)
}

func newHandshakePeerWithDefaults() *handshake {
	opts := handshakeOptions{
		Role: peer,
	}
	return newHandshake(newDefaultStrategy(), opts)
}

func genAlias() string {
	var aliasSlice []string
	npa := []string{
		"alfa", "bravo", "charlie", "delta", "echo", "foxtrot", "golf",
		"hotel", "india", "juliett", "kilo", "lima", "mike", "november",
		"oscar", "papa", "quebec", "romeo", "sierra", "tango", "uniform",
		"victor", "whiskey", "x-ray", "yankee", "zulu",
	}
	for i := 0; i < 3; i++ {
		x, _ := rand.Int(rand.Reader, big.NewInt(int64(len(npa))))
		aliasSlice = append(aliasSlice, npa[x.Int64()])
	}
	return strings.Join(aliasSlice, "-")
}

func newNegotiatorFromConfig(config peerConfig) (n negotiator, err error) {
	if n.Entropy, err = base64.StdEncoding.DecodeString(config.Entropy); err != nil {
		return
	}
	if n.Strategy, err = strategyFromConfig(config.Config); err != nil {
		return
	}
	n.Alias = config.Alias
	n.SortOrder = config.Item
	return
}

// TODO:
// - Get Session to support new Handshake (and store state)
// - Get Session to be able to receive peer bytes and import it into the handshake
// - Initiator should stay open until a complete task has been processed
// - Initiator should return its config as well as the others (if more than 2, we won't do that initially)
// - Initiator should have a total code count in his scanner
// - Once peer has scanned all codes needed, both parties should generate preshared keys
// - this will initiate setting up the actual chat
