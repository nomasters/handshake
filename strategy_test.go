package handshake

import (
	"encoding/base64"
	"testing"
)

func TestExportStrategy(t *testing.T) {
	privateKeyString := "6zjTWCoDkKESjroDj26qrw0/xSU0B14Co/lIZZHhbHUFFt6rMcqyLt21y1PmoPJbokhXrvO4p+zauvk+GuujzA=="
	privateKey, err := base64.StdEncoding.DecodeString(privateKeyString)
	if err != nil {
		t.Errorf("base64 failed to decode: %v\n", err)
	}
	publicKey := privateKey[32:]

	n := node{
		URL: "https://prototype.hashmap.sh",
	}

	sig := signatureAlgorithm{
		Type:       ed25519,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}

	rOpts := StorageOptions{
		WriteNodes: []node{n},
		Signatures: []signatureAlgorithm{sig},
		WriteRule:  defaultConsensusRule,
	}
	r, err := newHashmapStorage(rOpts)
	if err != nil {
		t.Errorf("new hashmapStore failed: %v\n", err)
	}

	settings := make(map[string]string)
	settings["query_type"] = "api"

	n2 := node{
		URL:      "https://ipfs.infura.io:5001/",
		Settings: settings,
	}
	sOpts := StorageOptions{
		WriteNodes: []node{n2},
		WriteRule:  defaultConsensusRule,
	}
	s, err := newIPFSStorage(sOpts)
	if err != nil {
		t.Errorf("new IPFS storage failed: %v\n", err)
	}

	c := newDefaultSBCipher()

	strat := strategy{
		Rendezvous: r,
		Storage:    s,
		Cipher:     c,
	}

	t.Log(strat.Config())
	stratJson, err := strat.ConfigJSONBytes()
	if err != nil {
		t.Errorf("failed on json bytes %v", err)
	}
	t.Log(string(stratJson))
}
