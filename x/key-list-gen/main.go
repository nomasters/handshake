package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"

	// "fmt"
	"encoding/gob"
	"io/ioutil"
)

func main() {
	m := make(map[string]string)
	g := make(map[string]string)
	for i := 0; i < 100000; i++ {
		m[genRandBase64(24)] = genRandBase64(32)
		g[genRandBase64(24)] = genRandBase64(32)
	}
	mb, _ := json.Marshal(m)

	err := ioutil.WriteFile("output.json", mb, 0644)
	if err != nil {
		panic(err)
	}

	gobFile, err := os.OpenFile("output.gob", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		panic(err)
	}
	defer gobFile.Close()
	enc := gob.NewEncoder(gobFile)
	if err := enc.Encode(g); err != nil {
		panic(err)
	}

}

func genRand(i int) []byte {
	key := make([]byte, i)
	rand.Read(key)
	return key
}

func genRandHex(i int) string {
	key := make([]byte, i)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func genRandBase64(i int) string {
	key := make([]byte, i)
	rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}
