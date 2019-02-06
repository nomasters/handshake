package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/nomasters/handshake"
	"github.com/nomasters/handshake/gen"
	"golang.org/x/crypto/blake2b"
)

func main() {
	password := "this really works"

	if err := checkOrCreateProfile(password); err != nil {
		log.Fatal(err)
	}

	ips := getUnicastIPs()
	fmt.Println(ips)

	session, err := handshake.NewDefaultSession(password)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	fmt.Println(string(session.GetProfile().ToJSON()))

	cert, err := gen.NewRSACerts()

	fmt.Println(cert.Certificate[0])
	hash := blake2b.Sum256(cert.Certificate[0])
	fmt.Printf("%x\n", hash)

	if err != nil {
		log.Fatal(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", helloHandler)

	srv := &http.Server{
		TLSConfig:    cfg,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		Addr:         ":8081",
	}

	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func checkOrCreateProfile(password string) error {
	exists, err := handshake.ProfilesExist()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return handshake.NewGenesisProfile(password)
}

func getUnicastIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.IsGlobalUnicast() {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}
