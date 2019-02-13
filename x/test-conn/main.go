package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"

	"golang.org/x/crypto/blake2b"
)

func main() {

	conn, err := tls.Dial("tcp", "127.0.0.1:8081", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	connState := conn.ConnectionState()
	pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: connState.PeerCertificates[0].Raw})
	conn.Close()

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(pem)
	if !ok {
		panic("failed to parse root certificate")
	}
	conn2, err := tls.Dial("tcp", "127.0.0.1:8081", &tls.Config{RootCAs: roots})
	if err != nil {
		panic("failed to connect: " + err.Error())
	}

	hash := blake2b.Sum256(connState.PeerCertificates[0].Raw)
	fmt.Printf("%x\n", hash)
	fmt.Println("it worked!")
	conn2.Close()

	fmt.Println(rand.Int(rand.Reader, big.NewInt(400)))
}
