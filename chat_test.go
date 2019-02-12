package handshake

import (
	"encoding/json"
	"testing"
)

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
