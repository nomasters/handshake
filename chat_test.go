package handshake

import (
	"encoding/json"
	"os"
	"testing"
)

func ensureBobCleanDB() {
	os.Remove("./bob-handshake.boltdb")
}

func ensureAliceCleanDB() {
	os.Remove("./alice-handshake.boltdb")
}

func TestNewChat(t *testing.T) {
	// ensureCleanDB()
	// defer ensureCleanDB()
	password := "hello,world"
	// if err := NewGenesisProfile(password); err != nil {
	// 	t.Fatal(err)
	// }
	s, err := NewDefaultSession(password)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.NewInitiatorWithDefaults()
	h := newHandshakePeerWithDefaults()
	p, err := h.Position.Share()
	if err != nil {
		t.Error(err)
	}
	if _, err := s.AddPeerToHandshake(p); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetHandshakePeerConfig(1); err != nil {
		t.Fatal(err)
	}
	_, err = s.NewChat()
	if err != nil {
		t.Fatal(err)
	}
	listBytes, err := s.ListChats()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(listBytes))

	var list []string

	json.Unmarshal(listBytes, &list)

	m := []byte(`{ "message": "hello, world" }`)

	// response, err := s.SendMessage(list[0], m)
	response, err := s.SendMessage("7a6d53b3a7b54823d91ea9b7", m)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(response))

}

func TestSendReceive(t *testing.T) {
	ensureCleanDB()
	ensureBobCleanDB()
	ensureAliceCleanDB()
	bobPassword := "so_damn_secure"
	bobStoragePath := "bob-handshake.boltdb"
	alicePassword := "youll_never_guess"
	aliceStoragePath := "alice-handshake.boltdb"

	if err := NewGenesisProfile(bobPassword); err != nil {
		t.Fatal(err)
	}

	os.Rename(defaultBoltFilePath, bobStoragePath)

	if err := NewGenesisProfile(alicePassword); err != nil {
		t.Fatal(err)
	}

	os.Rename(defaultBoltFilePath, aliceStoragePath)

	bobSesionOpts := SessionOptions{
		StorageEngine:   defaultStorageEngine,
		StorageFilePath: bobStoragePath,
	}

	bobSession, err := NewSession(bobPassword, bobSesionOpts)
	if err != nil {
		t.Fatal(err)
	}

	aliceSesionOpts := SessionOptions{
		StorageEngine:   defaultStorageEngine,
		StorageFilePath: aliceStoragePath,
	}

	aliceSession, err := NewSession(alicePassword, aliceSesionOpts)
	if err != nil {
		t.Fatal(err)
	}

	bobSession.NewInitiatorWithDefaults()
	aliceSession.NewPeerWithDefaults()

	aliceShare, err := aliceSession.ShareHandshakePosition()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := bobSession.AddPeerToHandshake(aliceShare); err != nil {
		t.Fatal(err)
	}

	bobShare, err := bobSession.GetHandshakePeerConfig(1)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := aliceSession.AddPeerToHandshake(bobShare); err != nil {
		t.Fatal(err)
	}

	bobChatID, err := bobSession.NewChat()
	if err != nil {
		t.Fatal(err)
	}

	aliceChatID, err := aliceSession.NewChat()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(bobChatID)
	t.Log(aliceChatID)

	m := []byte(`{ "message": "hello, world" }`)
	response, err := bobSession.SendMessage(bobChatID, m)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(response))

	m2 := []byte(`{ "message": "I'm also sending this second message." }`)
	response1, err := bobSession.SendMessage(bobChatID, m2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(response1))

	response2, err := aliceSession.RetrieveMessages(aliceChatID)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(response2))
}
