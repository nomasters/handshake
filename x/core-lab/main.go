package main

import (
	"fmt"
	"log"

	"github.com/nomasters/handshake"
)

func main() {

	primaryPassword := "this really works"
	// duressPassword := "oh, shit."

	// err = handshake.NewProfilesSetup(primaryPassword, duressPassword, storage)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	opts := handshake.SessionOptions{StorageEngine: handshake.BoltEngine}

	// Primary
	session, err := handshake.NewSession(primaryPassword, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	fmt.Println(string(session.Profile.ToJSON()))
}
