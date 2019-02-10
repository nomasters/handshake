package handshake

import (
	"os"
	"testing"
)

func ensureCleanDB() {
	os.Remove("./handshake.boltdb")
}

func TestNewDefaultSession(t *testing.T) {
	ensureCleanDB()
	defer ensureCleanDB()
	password := "hello,world"
	if err := NewGenesisProfile(password); err != nil {
		t.Fatal(err)
	}
	if s, err := NewDefaultSession(password); err != nil {
		t.Error(err)
	} else {
		s.Close()
	}
	s2, err := NewDefaultSession("junk password")
	if err == nil {
		t.Error("newDefaultSession returned no error with bad password")
	}
	if s2 != nil {
		s2.Close()
	}
}

func TestNewChat(t *testing.T) {
	ensureCleanDB()
	defer ensureCleanDB()
	password := "hello,world"
	if err := NewGenesisProfile(password); err != nil {
		t.Fatal(err)
	}
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
}
