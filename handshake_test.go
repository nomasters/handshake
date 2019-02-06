package handshake

import (
	"encoding/json"
	"testing"
)

func TestNewHandshake(t *testing.T) {
	h := newHandshakePeerWithDefaults()
	p, err := h.Position.Share()
	if err != nil {
		t.Errorf("sharing position failed: %v", err)
	}
	t.Log(string(p))

	var pc peerConfig

	if err := json.Unmarshal(p, &pc); err != nil {
		t.Error(err)
	}

	h2 := newHandshakeInitiatorWithDefaults()
	if err := h2.AddPeer(pc); err != nil {
		t.Error(err)
	}
	t.Log(h2)
}
