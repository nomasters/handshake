package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	totalKeys        = 100000
	lookupHashLength = 24
	oneTimeKeyLength = 32
	loadKeyChunk     = 10000
)

// LookupSet is a struct that holds byte slices related to lookup
// hashes and one time keys
type LookupSet struct {
	Lookup     []byte
	MAC        []byte
	OTPkey     []byte
	CipherText []byte
}

func main() {
	bucket := "lookups"

	// Get key from salt + password
	// this simulates time to get profile key
	pw := []byte("oh, hello, so, secure")
	salt := []byte("12238102398120318031238")
	profileKey := getKey(pw, salt)

	// initialize the bolt db with lord satan's permissions
	// and ensure the test bucket exists
	db, err := bolt.Open("db.bolt", 0666, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	if err := createBucket(db, bucket); err != nil {
		panic(err)
	}

	// generate list of lookups (used by loadKeys)
	m := genLookups(profileKey)
	// export LookupSets as map[string]string for JSON
	mmap := LookupSetsToMap(m)
	// takes map[string]string and a profile key and returns
	// and encrypted ciphertext as a byte slice
	blob := MarshallAndEncrypt(mmap, profileKey)

	// write the blob to storage
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		b.Put([]byte("encrypted_hashes"), blob)
		return nil
	})

	// take the []LookupSet and batch writes in loadKeyChunk sets
	loadKeys(db, m, bucket)

	// count the keys in the bucket
	countKeys(db, bucket)

	// if err := deleteBucket(db, bucket); err != nil {
	// 	panic(err)
	// }

}

// MarshallAndEncrypt takes a map[string]string and a profieKey and returns a byte slice.
// First the data is marshalled as a JSON byte slice and that data is fed into secret box
// and the ciphertext is returned
func MarshallAndEncrypt(m map[string]string, profileKey []byte) []byte {
	data, _ := json.Marshal(m)

	var sk [32]byte
	copy(sk[:], profileKey)
	var nonce [24]byte
	rand.Read(nonce[:])

	return secretbox.Seal(nonce[:], data, &nonce, &sk)
}

// LookupSetsToMap takes a slice of LookupSet and flattens the struct to a map[string]string.
// This map only includes the lookup and otpkey unencrypted, because it will be encrypted after
// marshalling. The source data byte slices are converted to base64 strings before being added
// the map.
func LookupSetsToMap(sets []LookupSet) map[string]string {
	m := make(map[string]string)
	for _, l := range sets {
		k := base64.StdEncoding.EncodeToString(l.Lookup)
		v := base64.StdEncoding.EncodeToString(l.OTPkey)
		m[k] = v
	}
	return m
}

// countKeys is a simple counter for keys in a bucket
// this is just a little helper func to see a count of what was written
func countKeys(db *bolt.DB, name string) {
	counter := 0
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}
		return nil
	})
	fmt.Println(counter)
}

// loadKeys takes a db pointer, lookupset slice  and bucket name and writes
// the MAC and CipherText of each lookupset as key and value to boltdb
// due to the volume of loaded data, this function breaks up the data loading into
// loadKeyChunk sized chunks to maximize performance
func loadKeys(db *bolt.DB, m []LookupSet, bucket string) {
	var keys []LookupSet
	i := 0

	for i < len(m) {
		if (i + loadKeyChunk) < len(m) {
			keys = m[i : i+loadKeyChunk]
			i = i + loadKeyChunk
		} else {
			keys = m[i:]
			i = len(m)
		}
		db.Update(func(tx *bolt.Tx) error {
			for _, v := range keys {
				b := tx.Bucket([]byte(bucket))
				b.Put(v.MAC, v.CipherText)
			}
			return nil
		})
	}
}

// createBucket is a helper function for creating a bucket if it doesn't already exist
func createBucket(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

// deleteBucket is a helper function for deleting a bucket
func deleteBucket(db *bolt.DB, name string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(name))
		if err != nil {
			return fmt.Errorf("delete bucket: %s", err)
		}
		return nil
	})
	return err
}

// sane defaults on argon2 key generation
func getKey(pw, salt []byte) []byte {
	return argon2.IDKey(pw, salt, 1, 64*1024, 4, 32)
}

// genLookups takes a profileKey and returns a slice of LookupSets
// this generates lookup and otpkey random data and generates MAC
// and ciphertext for bulk key generation
func genLookups(profileKey []byte) []LookupSet {

	var sbKey [32]byte
	copy(sbKey[:], profileKey)

	m := make([]LookupSet, totalKeys)
	for i := 0; i < totalKeys; i++ {
		lookup := genRand(lookupHashLength)
		otpkey := genRand(oneTimeKeyLength)
		h, _ := blake2b.New256(profileKey)
		mac := h.Sum(lookup)
		var nonce [24]byte
		rand.Read(nonce[:])

		data := make([]byte, lookupHashLength+oneTimeKeyLength)
		copy(data, lookup)
		copy(data[lookupHashLength:], otpkey)

		cipherText := secretbox.Seal(nonce[:], data, &nonce, &sbKey)

		m[i].Lookup = lookup
		m[i].MAC = mac
		m[i].OTPkey = otpkey
		m[i].CipherText = cipherText

	}
	return m
}

// genRand returns a byte slice of size i using `crypto/rand` data
func genRand(i int) []byte {
	key := make([]byte, i)
	rand.Read(key)
	return key
}
