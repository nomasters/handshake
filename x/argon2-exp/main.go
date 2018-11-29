package main

import (
	// "bufio"
	"crypto/rand"
	"fmt"
	// "os"
	"syscall"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	SeedLength = 12
)

// TODO
// - add secretbox encryption and decryption functions
// - add file save that uses seed as file name
// - add file open that uses passphrase + seed to gen key to open

func main() {

	pw := getPassword()
	salt := []byte("so damn salty")

	start := time.Now()
	key := getKey(pw, salt)
	t := time.Now()
	elapsed := t.Sub(start)

	fmt.Printf("key: %x generated in %s\n", key, elapsed)

}

func genSeed() []byte {
	k := make([]byte, SeedLength)
	rand.Read(k)
	return k
}

func getPassword() []byte {
	fmt.Print("Enter your password: ")
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	return pw
}

func getKey(pw, salt []byte) []byte {
	return argon2.IDKey(pw, salt, 1, 64*1024, 4, 32)
}
