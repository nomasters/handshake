package handshake

import (
	"bytes"
	"testing"
)

func TestGenLookups(t *testing.T) {
	var pepper [64]byte
	var entropy [96]byte

	copy(pepper[:], []byte("maich3zu1theeKahThi0CaechahZ1nei1ahcaitah1Au5quie5bee6PaeW5hie3y"))
	copy(entropy[:], []byte("aiphaiyu3aem2ko4ni4ohxohca1Iech9ohpie9uo9uij4Fe7hieVaowieh9ahGhiezeeyahZu9eeSahphaxaecaisutu0uij"))

	l1, err := genLookups(pepper, entropy, SecretBox, 100000)
	if err != nil {
		t.Error(err)
	}
	l2, err := genLookups(pepper, entropy, SecretBox, 100000)
	if err != nil {
		t.Error(err)
	}
	for k, v1 := range l1 {
		v2 := l2[k]
		if !bytes.Equal(v1, v2) {
			t.Error(err)
		}
	}
}
