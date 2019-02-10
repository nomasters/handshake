package handshake

import "encoding/json"

type strategy struct {
	Rendezvous storage
	Storage    storage
	Cipher     cipher
}

// strategyPeerConfig is a a struct that encapsulates the shared strategy settings for handshake
type strategyPeerConfig struct {
	Rendezvous peerStorage `json:"rendezvous"`
	Storage    peerStorage `json:"storage"`
	Cipher     peerCipher  `json:"cipher"`
}

// strategyConfig is a struct that encapsulates internal chat strategy settings
type strategyConfig struct {
	Rendezvous storageConfig
	Storage    storageConfig
	Cipher     cipherConfig
}

// Share returns the strategyPeerConfig for the strategy
func (s strategy) Share() (config strategyPeerConfig, err error) {
	if config.Rendezvous, err = s.Rendezvous.share(); err != nil {
		return
	}
	if config.Storage, err = s.Storage.share(); err != nil {
		return
	}
	if config.Cipher, err = s.Cipher.share(); err != nil {
		return
	}
	return
}

// Share returns the strategyPeerConfig for the strategy
func (s strategy) Export() (config strategyConfig, err error) {
	if config.Rendezvous, err = s.Rendezvous.export(); err != nil {
		return
	}
	if config.Storage, err = s.Storage.export(); err != nil {
		return
	}
	if config.Cipher, err = s.Cipher.export(); err != nil {
		return
	}
	return
}

func strategyFromPeerConfig(config strategyPeerConfig) (s strategy, err error) {
	if s.Rendezvous, err = newStorageFromPeer(config.Rendezvous); err != nil {
		return
	}
	if s.Storage, err = newStorageFromPeer(config.Storage); err != nil {
		return
	}
	if s.Cipher, err = newCipherFromPeer(config.Cipher); err != nil {
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

// ConfigJSONBytes marshall's the strategyPeerConfig as a json file
func (s strategy) ShareJSONBytes() ([]byte, error) {
	config, err := s.Share()
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
