package handshake

import "encoding/json"

type strategy struct {
	Rendezvous storage
	Storage    storage
	Cipher     cipher
}

// strategyConfig is a a struct that encapsulates the shared strategy settings for handshake
type strategyConfig struct {
	Rendezvous peerStorage `json:"rendezvous"`
	Storage    peerStorage `json:"storage"`
	Cipher     peerCipher  `json:"cipher"`
}

// Config returns the PeerConfig for the strategy
func (s strategy) Config() (config strategyConfig, err error) {
	if config.Rendezvous, err = s.Rendezvous.config(); err != nil {
		return
	}
	if config.Storage, err = s.Storage.config(); err != nil {
		return
	}
	if config.Cipher, err = s.Cipher.config(); err != nil {
		return
	}
	return
}

func strategyFromConfig(config strategyConfig) (s strategy, err error) {
	if s.Rendezvous, err = newStorageFromConfig(config.Rendezvous); err != nil {
		return
	}
	if s.Storage, err = newStorageFromConfig(config.Storage); err != nil {
		return
	}
	if s.Cipher, err = newCipherFromConfig(config.Cipher); err != nil {
		return
	}
	return
}

// ConfigJSONBytes marshall's the strategyConfig as a json file
func (s strategy) ConfigJSONBytes() ([]byte, error) {
	config, err := s.Config()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(config)
}

func newDefaultStrategy() strategy {
	return strategy{
		Rendezvous: newDefaultRendezvous(),
		Storage:    newDefaultMessageStorage(),
		Cipher:     newDefaultCipher(),
	}
}
