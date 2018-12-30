package main

import (
	"fmt"
	"log"

	"github.com/nomasters/handshake"
)

func main() {
	primaryPassword := "this really works"
	duressPassword := "oh, shit."

	if err := initProfiles(primaryPassword, duressPassword); err != nil {
		log.Fatal(err)
	}

	opts := handshake.SessionOptions{StorageEngine: handshake.BoltEngine}
	session, err := handshake.NewSession(primaryPassword, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	fmt.Println(string(session.Profile.ToJSON()))
}

func initProfiles(p, d string) error {
	opts := handshake.StorageOptions{Engine: handshake.DefaultStorageEngine}
	storage, err := handshake.NewStorage(opts)
	if err != nil {
		return err
	}
	defer storage.Close()
	cipher := handshake.NewTimeSeriesSBCipher()
	exists, err := handshake.ProfilesExist(storage)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return handshake.NewProfilesSetup(p, d, cipher, storage)
}
