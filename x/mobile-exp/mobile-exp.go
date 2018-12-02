package handshakeCore

import (
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/crypto/argon2"
)

// checking out simple string response
func HelloWorld() string {
	return "hello, world"
}

// testing out state in the app with a simple counter
type Counter struct {
	internal int
}

func (c *Counter) Inc()    { c.internal++ }
func (c Counter) Get() int { return c.internal }
func NewCounter() *Counter { return &Counter{} }

// testing network communication with a simple interface
// and error handling
func GetHash(hash string) ([]byte, error) {
	resp, err := http.Get("https://prototype.hashmap.sh/" + hash)
	if err != nil {
		return nil, errors.New("get request failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("body read failed")
	}
	return body, nil
}

// testing interacting with a more complex library
// but with a simple func interface
func GetKey(pw, salt []byte) []byte {
	return argon2.IDKey(pw, salt, 1, 64*1024, 4, 32)
}
