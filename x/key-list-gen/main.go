package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

// type

func main() {
	h := blake2b.Sum256([]byte("hello, world"))
	fmt.Printf("%x\n", h)
	fmt.Println(GenRandHex(64))
}

func GenRandHex(i int) string {
	key := make([]byte, i)
	rand.Read(key)
	return hex.EncodeToString(key)
}

// TODO

// generate key
// generate offset
// generate
