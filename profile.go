package handshake

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ProfileType is an enum type used by profiles
type ProfileType int

const (
	// ProfileIDLength is the length, in bytes, of the Profile.
	// This is used as both a unique identifier as well as the argon2 KDF salt
	ProfileIDLength = 24
	// ProfileKeyLength is the length of the key in bytes to by used by SecretBox
	ProfileKeyLength = 32
	// ProfileKeyPrefix is the prefix used for the profile keys
	ProfileKeyPrefix = "profiles/"
)

const (
	// PrimaryProfile is the profile type used by a logged in user
	PrimaryProfile ProfileType = iota
	// DuressProfile is the profile type used by dummy-duress accounts
	DuressProfile
)

// Profile represents a profile that has been accessed
// this would contain successfully decrypted profile data
type Profile struct {
	ID        string             `json:"id"`
	Type      ProfileType        `json:"type,string"`
	Key       string             `json:"key"`
	Delegated []DelegatedProfile `json:"delegated"`
	Settings  ProfileSettings    `json:"settings"`
}

// ProfileSettings holds profile settings info
type ProfileSettings struct {
	SessionTTL int64 `json:"sessionTTL"`
}

// DelegatedProfile holds delegated profile information
type DelegatedProfile struct {
	ID   string      `json:"id"`
	Type ProfileType `json:"type,string"`
	Key  string      `json:"key"`
}

// NewProfilesSetup takes primary and duress passwords and a storage interface and creates primary and
// duress profiles with the duress profile configured as a DelegratedProfile for the primary. Each profile
// then encrypts its configuration with the argon2 KDF. This function returns an error if both passwords are
// the same, or if there is a problem with the Storage interface.
func NewProfilesSetup(primaryPassword, duressPassword string, cipher Cipher, storage Storage) error {
	profiles, err := storage.List(ProfileKeyPrefix)
	if err != nil {
		return err
	}
	if len(profiles) > 0 {
		return errors.New("existing profiles found: this function may only be used for initial setup")
	}
	if primaryPassword == "" {
		return errors.New("invalid password length for primary")
	}
	if duressPassword == "" {
		return errors.New("invalid password length for duress")
	}
	if primaryPassword == duressPassword {
		return errors.New("primary and duress passwords must differ")
	}

	primary := GenerateProfile(PrimaryProfile)
	duress := GenerateProfile(DuressProfile)
	primary.AddDelegate(duress)

	if err := initProfile(primary, primaryPassword, cipher, storage); err != nil {
		return err
	}
	return initProfile(duress, duressPassword, cipher, storage)
}

func initProfile(p Profile, password string, cipher Cipher, storage Storage) error {
	key := DeriveKey([]byte(password), p.IDBytes())

	data, err := cipher.Encrypt(p.ToJSON(), key)
	if err != nil {
		return err
	}
	return storage.Set(ProfileKeyPrefix+p.ID, data)
}

// GetProfileFromEncryptedStorage takes a storage path, password, and storage interface and returns a Profile struct and error.DecryptProfile
//
func GetProfileFromEncryptedStorage(path string, key []byte, cipher Cipher, storage Storage) (Profile, error) {
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
	saltB64 := strings.Replace(path, ProfileKeyPrefix, "", 1)
	id, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// GenerateProfile take a ProfileType and returns a Profile struct with a randomly generated ID and key
func GenerateProfile(profileType ProfileType) Profile {
	return Profile{
		ID:   base64.StdEncoding.EncodeToString(genRandBytes(ProfileIDLength)),
		Type: profileType,
		Key:  base64.StdEncoding.EncodeToString(genRandBytes(ProfileKeyLength)),
	}
}

// AddDelegate takes a Profile and adds it to the current profile as a DelegatedProfile
func (p *Profile) AddDelegate(d Profile) {
	dp := DelegatedProfile{
		ID:   d.ID,
		Type: d.Type,
		Key:  d.Key,
	}
	p.Delegated = append(p.Delegated, dp)
}

// IDBytes converts the ID string in base64 to decoded bytes
func (p Profile) IDBytes() []byte {
	id, _ := base64.StdEncoding.DecodeString(p.ID)
	return id
}

// ToJSON is a helper method to make it easier to convert Profile to JSON
func (p Profile) ToJSON() []byte {
	pb, _ := json.Marshal(p)
	return pb
}

// String method converts enum type into a human readable format
func (p ProfileType) String() string {
	profiles := []string{"primary", "duress"}
	return profiles[p]
}

// MarshalJSON encodes ProfileType in human readable format for JSON
func (p ProfileType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, p)), nil
}

// UnmarshalJSON decodes human readable format to enum type
func (p *ProfileType) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case "primary":
		*p = PrimaryProfile
		return nil
	case "duress":
		*p = DuressProfile
		return nil
	}
	return errors.New("invalide profile type: " + string(b))
}
