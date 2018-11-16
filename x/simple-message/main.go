/*

The purpose of this is to experiment with the message primative of posting and reteriveing messages
using ipfs and hashmap.

This toy app assumes you are using a file called lookup.json which can be generated  in the x/key-list-gen dir.

*/

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-ipfs-api"
	"github.com/nomasters/hashmap"
	"golang.org/x/crypto/nacl/secretbox"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	defaultIPFSRPCPath       = "https://ipfs.infura.io:5001"
	defaultChunkSize         = 16000
	defaultMessageTTL        = 86400 // one day
	defaultChatPayloadMethod = "nacl-secretbox-16000"
	defaultHashmapEndpoint   = "https://prototype.hashmap.sh"
)

type ChatPayload struct {
	Method string `json:"method"`
	Lookup string `json:"lookup"`
	Data   string `json:"data"`
}

type ChatData struct {
	Parent    string   `json:"parent"`
	Timestamp int64    `json:"timestamp"`
	Media     []string `json:"media"`
	Message   string   `json:"message"`
	TTL       int      `json:"ttl"`
}

type ChatParticipant struct {
	PersonalLookup map[string]string
	OtherLookup    map[string]string
	PrivateKey     []byte
	ParentHash     string
	OtherPubKey    []byte
}

type HashmapResponse struct {
	Endpoint string `json:"endpoint"`
}

func main() {

	// initialize bob and alice identities and lookup keys

	fmt.Println("initializing")

	alice := NewChatParticipant()
	bob := NewChatParticipant()
	ExchangeKeys(&alice, &bob)

	// generate and submit first message from alice

	fmt.Println("\ngenerating message from alice to bob")
	fmt.Println("------------------------------------\n")

	originalMessage := "hello, world"

	chatPayload := alice.MakeChatPayload(originalMessage)
	fmt.Println(string(chatPayload))

	fmt.Println("\nsubmitting message from alice to ipfs")
	fmt.Println("-------------------------------------\n")
	alice.SubmitToIPFS(chatPayload)
	fmt.Println("ipfs hash: " + alice.ParentHash)

	fmt.Println("\ngenerate hashmap payload for message")
	fmt.Println("------------------------------------\n")

	hp := alice.MakeHashmapPayload(alice.ParentHash)

	fmt.Println(string(hp))

	fmt.Println("\nsubmitting hashmap payload")
	fmt.Println("--------------------------\n")

	endpoint := SubmitPayloadToHashmap(hp)

	fmt.Println("hashmap endpoint: " + endpoint)

	fmt.Println("\nmessage submission complete\n")

	fmt.Println("\nretreving message from alice to bob: checking alice's hashmap endpoint")
	fmt.Println("----------------------------------------------------------------------\n")

	fmt.Println("\nretreving message from alice to bob: retreiving encrypted ipfs data")
	fmt.Println("-------------------------------------------------------------------\n")

	// retreive message as bob from alice
	payload := bob.GetIPFSDataFromHashmapEndpoint(endpoint)
	fmt.Println(string(payload))

	fmt.Println("\nretreving message from alice to bob: decrypting message from payload")
	fmt.Println("--------------------------------------------------------------------\n")

	message := bob.DecryptMessageFromPayload(payload)
	fmt.Println(string(message))
}

func (p *ChatParticipant) DecryptMessageFromPayload(payload []byte) []byte {

	var cp ChatPayload
	if err := json.Unmarshal(payload, &cp); err != nil {
		panic(err)
	}

	var key [32]byte
	k, _ := base64.StdEncoding.DecodeString(p.OtherLookup[cp.Lookup])
	copy(key[:], k)

	db, _ := base64.StdEncoding.DecodeString(cp.Data)
	decryptedData := decrypt(db, key)
	delete(p.OtherLookup, cp.Lookup)

	return []byte(decryptedData)
}

func (p *ChatParticipant) GetIPFSDataFromHashmapEndpoint(endpoint string) []byte {
	hmp := GetPayloadFromHashmap(endpoint)
	pl, err := hashmap.NewPayloadFromReader(bytes.NewReader(hmp))
	if err != nil {
		panic(err)
	}
	d, _ := pl.GetData()
	hashmapMessage, _ := d.MessageBytes()
	var hashMapPayload ChatPayload
	if err := json.Unmarshal(hashmapMessage, &hashMapPayload); err != nil {
		panic(err)
	}

	var key [32]byte
	k, _ := base64.StdEncoding.DecodeString(p.OtherLookup[hashMapPayload.Lookup])
	copy(key[:], k)

	db, _ := base64.StdEncoding.DecodeString(hashMapPayload.Data)
	decryptedHashmapData := decrypt(db, key)
	delete(p.OtherLookup, hashMapPayload.Lookup)

	ipfsData := GetFromIPFS(string(decryptedHashmapData))
	// TODO: not sure why the first 7 bytes and last 5 are malformed.
	// neet do investigate
	return []byte(ipfsData[7 : len(ipfsData)-5])
}

func NewChatParticipant() ChatParticipant {
	return ChatParticipant{
		PersonalLookup: genLookup(),
		OtherLookup:    make(map[string]string),
		PrivateKey:     hashmap.GenerateKey(),
	}
}

// ExchangeKeys copies the lookupKey data between two ChatParticipants and copies the pubkey portion
// of an ed25519 privatekey. This is a gross oversimilificaiton of what will really happen, but is
// easy to configure and convenient for this POC
func ExchangeKeys(a, b *ChatParticipant) {
	for k := range a.PersonalLookup {
		b.OtherLookup[k] = a.PersonalLookup[k]
	}
	for k := range b.PersonalLookup {
		a.OtherLookup[k] = b.PersonalLookup[k]
	}
	a.OtherPubKey = b.PrivateKey[32:]
	b.OtherPubKey = a.PrivateKey[32:]
}

func genLookup() map[string]string {
	l := make(map[string]string)
	for i := 0; i < 20; i++ {
		l[GenRandBase64(12)] = GenRandBase64(32)
	}
	return l
}

func GetPayloadFromHashmap(endpoint string) []byte {
	resp, err := http.Get(defaultHashmapEndpoint + "/" + endpoint)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return body
}

func SubmitPayloadToHashmap(payload []byte) string {
	resp, err := http.Post(defaultHashmapEndpoint, "application/json", bytes.NewReader(payload))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var hmr HashmapResponse
	if err := json.Unmarshal(body, &hmr); err != nil {
		panic(err)
	}
	return hmr.Endpoint
}

func GenRandBase64(i int) string {
	key := make([]byte, i)
	rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}

func GetFromIPFS(endpoint string) string {
	sh := shell.NewShell(defaultIPFSRPCPath)
	obj, err := sh.ObjectGet("/ipfs/" + endpoint)
	if err != nil {
		panic(err)
	}
	return obj.Data
}

func (p *ChatParticipant) SubmitToIPFS(payload []byte) {
	sh := shell.NewShell(defaultIPFSRPCPath)
	mhash, err := sh.Add(bytes.NewReader(payload))
	if err != nil {
		panic(err)
	}
	// set parentHash for future messages
	p.ParentHash = mhash
}

func (p ChatParticipant) MakeHashmapPayload(message string) []byte {
	var lookup string
	var key [32]byte

	// quick and dirty grab of the latest key
	// not concurency safe
	for k := range p.PersonalLookup {
		lookup = k
		break
	}
	k, _ := base64.StdEncoding.DecodeString(p.PersonalLookup[lookup])
	copy(key[:], k)

	// delete the key/value from lookup after extracting
	delete(p.PersonalLookup, lookup)

	encryptedData := encrypt([]byte(message), key)

	cp := ChatPayload{
		Method: defaultChatPayloadMethod,
		Lookup: lookup,
		Data:   base64.StdEncoding.EncodeToString(encryptedData),
	}
	payloadMarshalled, _ := json.Marshal(cp)
	opts := hashmap.GeneratePayloadOptions{Message: string(payloadMarshalled)}
	payload, _ := hashmap.GeneratePayload(opts, p.PrivateKey)
	return payload
}

func (p ChatParticipant) MakeChatPayload(message string) []byte {
	var lookup string
	var key [32]byte

	data := ChatData{
		Parent:    p.ParentHash,
		Timestamp: time.Now().UnixNano(),
		Message:   message,
		TTL:       defaultMessageTTL,
	}

	dataJSON, _ := json.Marshal(data)

	// quick and dirty grab of the latest key
	// not concurency safe
	for k := range p.PersonalLookup {
		lookup = k
		break
	}
	k, _ := base64.StdEncoding.DecodeString(p.PersonalLookup[lookup])
	copy(key[:], k)

	// delete the key/value from lookup after extracting
	delete(p.PersonalLookup, lookup)

	encryptedData := encrypt(dataJSON, key)

	payload := ChatPayload{
		Method: defaultChatPayloadMethod,
		Lookup: lookup,
		Data:   base64.StdEncoding.EncodeToString(encryptedData),
	}
	payloadMarshalled, _ := json.Marshal(payload)
	return payloadMarshalled
}

func encrypt(data []byte, key [32]byte) []byte {
	var nonce [24]byte
	rand.Read(nonce[:])
	return secretbox.Seal(nonce[:], data, &nonce, &key)
}

func decrypt(data []byte, key [32]byte) []byte {
	var nonce [24]byte
	copy(nonce[:], data[:24])
	d, ok := secretbox.Open(nil, data[24:], &nonce, &key)
	if !ok {
		panic("decryption error")
	}
	return d
}
