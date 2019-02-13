package handshake

import (
	"encoding/base64"
	"testing"
)

func TestHashmapSet(t *testing.T) {
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

	opts := StorageOptions{
		WriteNodes: []node{n},
		Signatures: []signatureAlgorithm{sig},
		WriteRule:  defaultConsensusRule,
	}
	hms, err := newHashmapStorage(opts)
	if err != nil {
		t.Errorf("new hashmapStore failed: %v\n", err)
	}

	if _, err := hms.Set("", []byte("this was generated from a go test. right now for real")); err != nil {
		t.Errorf("set failed: %v\n", err)
	}
}

func TestHashmapStorageGet(t *testing.T) {
	n := node{
		URL: "https://prototype.hashmap.sh/2DrjgbL8QfKRvxU9KtFYFdNiPZrQijyxkvWXH17QnvNmzB3apR",
		//    2Drjgb5DseoVAvRLngcVmd4YfJAi3J1145kiNFV3CL32Hs6vzb
	}
	opts := StorageOptions{
		ReadNodes: []node{n},
		ReadRule:  defaultConsensusRule,
	}
	hms, err := newHashmapStorage(opts)
	if err != nil {
		t.Errorf("new hashmapStore failed: %v\n", err)
	}
	response, err := hms.Get("")
	if err != nil {
		t.Errorf("response failed: %v\n", err)
	}
	t.Log(string(response))
}

func TestGetFromIPFS(t *testing.T) {
	settings := make(map[string]string)
	settings["query_type"] = "api"

	happyNodes := []node{
		// node{
		// 	URL:      "http://127.0.0.1:5001",
		// 	Settings: settings,
		// },
		node{
			URL:      "https://ipfs.infura.io:5001/",
			Settings: settings,
		},
		node{
			URL: "https://cloudflare-ipfs.com",
		},
	}
	hash := "QmZULkCELmmk5XNfCgTnCyFgAVxBRBXyDHGGMVoLFLiXEN"
	for _, n := range happyNodes {
		resp, err := getFromIPFS(n, hash)
		if err != nil {
			t.Error(err)
		}
		t.Log(string(resp))
	}
}

func TestPostToIPFS(t *testing.T) {
	settings := make(map[string]string)

	settings["query_type"] = "api"

	happyNodes := []node{
		// node{
		// 	URL:      "http://127.0.0.1:5001",
		// 	Settings: settings,
		// },
		node{
			URL:      "https://ipfs.infura.io:5001/",
			Settings: settings,
		},
		node{
			URL: "https://hardbin.com",
		},
	}
	body := []byte("hello, world")
	for _, n := range happyNodes {
		resp, err := postToIPFS(n, body)
		if err != nil {
			t.Error(err)
		}
		t.Log(string(resp))
	}
}
