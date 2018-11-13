package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func main() {
	m := make(map[string]string)
	for i := 0; i < 100000; i++ {
		m[GenRandBase64(12)] = GenRandBase64(32)
	}
	mb, _ := json.Marshal(m)
	fmt.Println(string(mb))
}

func GenRandHex(i int) string {
	key := make([]byte, i)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func GenRandBase64(i int) string {
	key := make([]byte, i)
	rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}
