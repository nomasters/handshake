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
