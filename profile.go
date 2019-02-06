package handshake

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

const (
	// ProfileIDLength is the length, in bytes, of the Profile.
	// This is used as both a unique identifier as well as the argon2 KDF salt
	profileIDLength = 24
	// ProfileKeyLength is the length of the key in bytes to by used by SecretBox
	profileKeyLength = 32
	// profileKeyPrefix is the prefix used for the profile keys
	profileKeyPrefix = "profiles/"
)

// Profile represents a profile that has been accessed
// this would contain successfully decrypted profile data
type Profile struct {
	ID       string          `json:"id"`
	Key      string          `json:"key"`
	Settings profileSettings `json:"settings"`
}

// ProfileSettings holds profile settings info
type profileSettings struct {
	SessionTTL int64 `json:"sessionTTL"`
}

// ProfilesExist configures a storage engine and checks `profilesExist`. It returns a bool and error.
// This is used on app startup to check to see if this is the first time running the tool. If this function
// returns `false` and no errors, the next step would be to prompt the user to setup a new profile using
// `NewGenesisProfile()`.
func ProfilesExist() (bool, error) {
	opts := StorageOptions{Engine: defaultStorageEngine}
	storage, err := newStorage(opts)
	if err != nil {
		return false, err
	}
	defer storage.Close()
	return profilesExist(storage)
}

// ProfilesExist takes a storage interface and checks to see if any profiles exist
// returns true only if the list of profiles is greater than 0 all other failure states return false.
func profilesExist(storage storage) (bool, error) {
	profiles, err := storage.List(profileKeyPrefix)
	if err != nil {
		return false, err
	}
	if len(profiles) > 0 {
		return true, nil
	}
	return false, nil
}

// NewGenesisProfile takes password and
func NewGenesisProfile(password string) error {
	opts := StorageOptions{Engine: defaultStorageEngine}
	storage, err := newStorage(opts)
	if err != nil {
		return err
	}
	defer storage.Close()
	exists, err := profilesExist(storage)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("existing profiles found: this function may only be used for initial setup")
	}

	return initProfile(generateRandomProfile(), password, newTimeSeriesSBCipher(), storage)
}

func initProfile(p Profile, password string, cipher cipher, storage storage) error {
	key := deriveKey([]byte(password), p.IDBytes())

	data, err := cipher.Encrypt(p.ToJSON(), key)
	if err != nil {
		return err
	}
	_, err = storage.Set(profileKeyPrefix+p.ID, data)
	return err
}

// GetProfileFromEncryptedStorage takes a storage path, password, and storage interface and returns a Profile struct and error.DecryptProfile
//
func getProfileFromEncryptedStorage(path string, key []byte, cipher cipher, storage storage) (Profile, error) {
	var p Profile

	data, err := storage.Get(path)
	if err != nil {
		return p, err
	}

	pBytes, err := cipher.Decrypt(data, key)
	if err != nil {
		return p, err
	}

	if err := json.Unmarshal(pBytes, &p); err != nil {
		return p, err
	}

	return p, nil
}

func getIDFromPath(path string) ([]byte, error) {
	saltHex := strings.Replace(path, profileKeyPrefix, "", 1)
	id, err := hex.DecodeString(saltHex)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// GenerateRandomProfile returns a Profile struct with a randomly generated ID and key
func generateRandomProfile() Profile {
	return Profile{
		ID:  hex.EncodeToString(genRandBytes(profileIDLength)),
		Key: base64.StdEncoding.EncodeToString(genRandBytes(profileKeyLength)),
	}
}

// IDBytes converts the ID string in base64 to decoded bytes
func (p Profile) IDBytes() []byte {
	id, _ := hex.DecodeString(p.ID)
	return id
}

// ToJSON is a helper method to make it easier to convert Profile to JSON
func (p Profile) ToJSON() []byte {
	pb, _ := json.Marshal(p)
	return pb
}
