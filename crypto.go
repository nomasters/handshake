package handshake

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// SecretBoxDefaultChunkSize is the default size of an encrypted chunk of data
	SecretBoxDefaultChunkSize = 16000
	// SecretBoxDecryptionOffset is the additional offset of bytes needed to offset
	// for the nonce and authentication bytes
	SecretBoxDecryptionOffset = 40
	// SecretBoxNonceLength is the length in bytes required for the nonce
	SecretBoxNonceLength = 24
	// SecretBoxKeyLength is the length in bytes required for the key
	SecretBoxKeyLength = 32
)

// NonceType is used for type enumeration for Ciphers
type NonceType int

const (
	// RandomNonce is the NonceType used for pure crypto/rand generated nonces
	RandomNonce NonceType = iota
	// TimeSeriesNonce is the NonceType used for 4 byte unix time prefixed crypto/rand generated nonces
	TimeSeriesNonce
)

// Cipher is an interface used for encrypting and decrypting byte slices.
type Cipher interface {
	Encrypt(data []byte, key []byte) ([]byte, error)
	Decrypt(data []byte, key []byte) ([]byte, error)
}

// genRandBytes takes a length of l and returns a byte slice of random data
func genRandBytes(l int) []byte {
	b := make([]byte, l)
	rand.Read(b)
	return b
}

// genTimeStampNonce takes an int for the nonce size and returns a byte slice of length size.
// A byte slice is created for the nonce and filled with random data from `crypto/rand`, then the
// first 4 bytes of the nonce are overwritten with LittleEndian encoding of `time.Now().Unix()`
// The purpose of this function is to avoid an unlikely collision in randomly generating nonces
// by prefixing the nonce with time series data.
func genTimeStampNonce(l int) []byte {
	nonce := genRandBytes(l)
	timeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeBytes, uint64(time.Now().Unix()))
	copy(nonce, timeBytes[:4])
	return nonce
}

// DeriveKey takes a password and salt and applies a set of fixed parameters
// to the argon2 IDKey algorithm.
func DeriveKey(pw, salt []byte) []byte {
	return argon2.IDKey(pw, salt, 1, 64*1024, 4, SecretBoxKeyLength)
}

// SecretBoxCipher is a struct and method set that conforms to the Cipher interface. This is the primary cipher used
// for all blob encryption and decryption for handshake
type SecretBoxCipher struct {
	Nonce     NonceType
	ChunkSize int
}

// NewTimeSeriesSBCipher returns a timeSeriesNonce based SecretBoxCipher struct that conforms to the
// Cipher interface
func NewTimeSeriesSBCipher() SecretBoxCipher {
	return SecretBoxCipher{Nonce: TimeSeriesNonce, ChunkSize: SecretBoxDefaultChunkSize}
}

// NewDefaultSBCipher returns a RandomNonce based SecretBoxCipher struct that conforms to the
// Cipher interface
func NewDefaultSBCipher() SecretBoxCipher {
	return SecretBoxCipher{Nonce: RandomNonce, ChunkSize: SecretBoxDefaultChunkSize}
}

// Encrypt takes byte slices for data and a key and returns the ciphertext output for secretbox
func (s SecretBoxCipher) Encrypt(data []byte, key []byte) ([]byte, error) {
	var encryptedData []byte
	chunkSize := s.ChunkSize

	if len(key) != SecretBoxKeyLength {
		return encryptedData, errors.New("invalid key length")
	}

	var k [SecretBoxKeyLength]byte
	copy(k[:], key)

	for i := 0; i < len(data); i = i + chunkSize {
		var chunk []byte
		if len(data[i:]) >= chunkSize {
			chunk = data[i : i+chunkSize]
		} else {
			chunk = data[i:]
		}
		nonce := s.GenNonce()

		var n [SecretBoxNonceLength]byte
		copy(n[:], nonce)

		encryptedChunk := secretbox.Seal(n[:], chunk, &n, &k)
		encryptedData = append(encryptedData, encryptedChunk...)
	}
	return encryptedData, nil
}

// Decrypt takes byte slices for data and key and returns the clear text output for secretbox
func (s SecretBoxCipher) Decrypt(data []byte, key []byte) ([]byte, error) {
	var decryptedData []byte
	chunkSize := s.ChunkSize + SecretBoxDecryptionOffset

	if len(key) != SecretBoxKeyLength {
		return decryptedData, errors.New("invalid key length")
	}

	var k [SecretBoxKeyLength]byte
	copy(k[:], key)

	for i := 0; i < len(data); i = i + chunkSize {
		var chunk []byte
		if len(data[i:]) >= chunkSize {
			chunk = data[i : i+chunkSize]
		} else {
			chunk = data[i:]
		}
		var n [SecretBoxNonceLength]byte
		copy(n[:], chunk[:SecretBoxNonceLength])

		decryptedChunk, ok := secretbox.Open(nil, chunk[SecretBoxNonceLength:], &n, &k)
		if !ok {
			return nil, errors.New("decrypt failed")
		}
		decryptedData = append(decryptedData, decryptedChunk...)
	}
	return decryptedData, nil
}

// GenNonce returns a set of nonce bytes based on the NonceType configured in the struct
func (s SecretBoxCipher) GenNonce() []byte {
	switch s.Nonce {
	case RandomNonce:
		return genRandBytes(SecretBoxNonceLength)
	case TimeSeriesNonce:
		return genTimeStampNonce(SecretBoxNonceLength)
	default:
		return genRandBytes(SecretBoxNonceLength)
	}
}
