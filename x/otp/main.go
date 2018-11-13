package main

import (
	"crypto/rand"
	"fmt"
)

func main() {
	otp := gen()
	message := []byte(`This is my super secret message. It can be any length. Messages shorter than 512 bytes will be zero padded, and messages longer than 512 bytes will be chopped off at 512 bytes.`)

	m := enforce(message)
	c := xor(m, otp)
	d := xor(c, otp)

	fmt.Printf("source:\n%s\n", string(message))
	fmt.Printf("otp:\n%x\n", otp)
	fmt.Printf("cipher:\n%x\n", c)
	fmt.Printf("decrypt:\n%s\n", string(d[:]))

}

func gen() [512]byte {
	key := make([]byte, 512)
	rand.Read(key)
	return enforce(key)
}

func enforce(m []byte) (o [512]byte) {
	copy(o[:], m[:])
	return
}

func xor(m, k [512]byte) (o [512]byte) {
	for i, _ := range m {
		o[i] = m[i] ^ k[i]
	}
	return
}
