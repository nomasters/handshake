# Handshake Core Spec Doc

## Summary

This is the spec document for handshake-core, the underlying structure specification for the cli and GUI apps as well as a proposal on how data is structured and stored.

## Introduction

Handshake is designed to be a decentralized p2p encrypted communications tool based on in-person initialization of communication so that all future transmissions rely symmetric key cryptography. This is primarily a design for out-of-band communication in which communicating parties are worried about potential compromises in asymmetric encryption methodology, CA poisoning, or relying on centralized service providers for communications technology.

Handshake is designed initially to work on IPFS, but there are no technical reasons other backends couldn't be supported.

Unique characteristics of handshake:

- There are no centralized user accounts, no identifying information about the users handled by a third part. This is possible because each new chat session utilizes a uniquely generated ipfs config, which has a unique id and private key. This is used specifically for IPNS, to tag the "latest" update to a message history.
- encryption keys are never transmitted over the internet. This is due to the nature of the initial configuration.
- Meta-data about who is involved in the chat is never transmitted over the public internet.

Here is a scenario for initializing a handshake session:

- Bob taps on the new chat icon and is offered Initiate or Join. He chooses initiate.
- Alice taps on the new chat icon and is offered Initiate or Join. She chooses join. 
- Bob is presented with a couple of basic configuration questions and then a QR code is displayed
- Alice is prompted to scan the initiator code. 
- Bob and Alice's devices connect and each other and go through the process of sharing keys and configuration data
- An initialization message is posted from each device to ensure everything is setup properly
- bob and Alice have successfully setup a new handshake chat

Under the hood this is what is happening:

- When a new chat is initialized, either as an initiator or a joiner, a new IPFS configuration, which includes creation of an IPFS private key and Peer ID. this is the only PKI used on for handshake, its used to integrating with IPNS for mutable state of the chat. This means that each and every chat gets a new and unique PeerId and private key.
- Bob generates an initial symmetric key, along with the port and route for Alice to connect to, this initial data a json blob encoded into a QR code
- When Alice scans the code, the has all the necessary info to start the process which includes the IPFS peerId of bob, the encryption key for initializing handshake, and the other connection details.
- Alice replies to Bob by encrypting a json payload that includes her IPFS PeerId and her generated preshared keys.
- Bob replies with an encrypted payload of his preshared keys.
- Bob posts his first message to handshake by creating a new message payload and writing it to IPFS. This returns an immutable IPFS hash. Bob then pins this hash to IPNS with his PeerId and private key
- Alice does the same.
- Bob queries Alice's IPNS endpoint to get her message. 
- Alice does the same.
- Initialization is complete.

The generation of preshared keys involves generating a set of random lookup hashes and random keys. The default behavior is to generate some large number (possibly 100k) lookup hashes and keys, but this should be configurable by Bob to generate fewer (to force a shorter conversation).

```
{
	"e1f1534d26d2": e5ac434796dab5f5164796d4"
	...
}

```

These keys are used to create one-time use message encryption. I message payload looks as follows:

```
{
    "crypto": {
        "method": "secretbox",
        "chunk": 16000
    },
    "hash": "e1f1534d26d2",
    "payload": "YXNkcG9mOWpzYVswOWZ3ZlswMm5mIDIzMmw7M2tqZnc7ZWZsamFzO2RvaWZ1YXNqIGQ7bGZqYXNkbDtrZiBqYXNkO2ZqYXNkIGZhc2Rwb2Y5anNhWzA5ZndmWzAybmYgMjMybDsza2pmdztlZmxqYXM7ZG9pZnVhc2ogZDtsZmphc2RsO2tmIGphc2Q7Zmphc2QgZmFzZHBvZjlqc2FbMDlmd2ZbMDJuZiAyMzJsOzNramZ3O2VmbGphcztkb2lmdWFzaiBkO2xmamFzZGw7a2YgamFzZDtmamFzZCBmYXNkcG9mOWpzYVswOWZ3ZlswMm5mIDIzMmw7M2tqZnc7ZWZsamFzO2RvaWZ1YXNqIGQ7bGZqYXNkbDtrZiBqYXNkO2ZqYXNkIGZhc2Rwb2Y5anNhWzA5ZndmWzAybmYgMjMybDsza2pmdztlZmxqYXM7ZG9pZnVhc2ogZDtsZmphc2RsO2tmIGphc2Q7Zmphc2QgZg=="
}
```

It includes a description of the cryptography used. The initial implementation will use `secretbox` which requires a chunk size for larger messages to be processed efficiently. The the `hash` is the lookup hash that the message recipient would use to lookup which secret key to use. The payload is encoded with base64 encoding and follows the `secretbox` spec in which the chunk nonce and authentication bytes exists before each encrypted chunk in the payload.

Once decrypted, a payload would look something like this:

```
{
    "parent": [IPFS_HASH_OF_PREVIOUS_MESSAGE],
    "time_stamp": [UNIX_TIME_STAMP],
    "media": [
    	[BASE64_OF_MEDIA_ITEM],
    	[BASE64_OF_MEDIA_ITEM]
    ],
    "message": [BASE64_OF_MESSAGE],
    "ttl": 700,
    ....
}
```
The specific details are being worked out, but the primary structure is here.

- Each message references the `parent` message. This allows for Bob to update messages as often as he wants, once Alice gets the latest message, she can continue to query the parent message IPFS immutable hash until she's reached a message that contains a hash that she's already received. 
- `time_stamp` is optional but recommended, if no `time_stamp` is present, the app will used received time. This is used to help weave two IPNS conversation endpoints together
- a message can contain media and a body,
- `media` is 
- `message` is for the message body of the payload
- `ttl` is the TTL before the decrypted message is destroyed

Upon receiving a message and successfully decrypting the message, the key is destroyed.

The `lookupHash` is the index for various types of lookups:

- decryption key lookup - upon receiveing a new message, the lookupHash is used to find the decryption key. if the key hasn't been used, the `lookupHash` is used to get the key from the key table. Two possible error states are that they lookupHash isn't found or, it is found, but in the consumed lookupKey table, which which give a warning that a key was used twice. When a key is used and added to the consumed key list, it's entry should be removed from the lookup table. As cited above, the structure should look as follows:

```
{
	"e1f1534d26d2": e5ac434796dab5f5164796d4"
	...
}

```

NOTE: Bob and Alice each have their own lookupHash tables, Bob's was generated by Bob and Alice by Alice on initialization

- consumed hash - A consumed key is linked to a specific lookuphash + IPFS hash. the lookupKey + hash gives a immutable consumption hash pair. If a lookup key is ever used twice, the ipfs hashes will differ and the lookup should be rejected (and the recipient notified). The data structure should look as follows:

```
lookupHash  = e1f1534d26d2
ipfsHash    = QMUTwqtL5Veio86ZZ3u5TmGpTnM5iCA4bmTvbTVocUTo3N
messageHash = lookupHash + ipfsHash

{
	"e1f1534d26d2QMUTwqtL5Veio86ZZ3u5TmGpTnM5iCA4bmTvbTVocUTo3N": [UNIX_TIME_STAMP_WHEN_MESSAGE_CONSUMED],
	...
}
```

NOTE: the consumed key table isn't required. If this table is wiped, a failed match or duplicate hash are treated the same. This would allow Alice to periodically wipe usage metadata if she saw fit. As with the key lookup table, consumed keys are stored in unique tables as well.

- message history - The message history table is used to store decrypted messages locally on the device. This is a time-series based data feed that blends messages form the sender and recipient in a single view.

```
{
	"messages": [
		{
			'messageId':[MESSAGE_HASH],
			'senderId': [IPFS_PeerID],
			'sentAt': [UNIX_TIME_STAMP],
			'receivedAt': [UNIX_TIME_STAMP],
			'expiration': [UNIX_TIME_STAMP],
			'payload': [decrypted_payload_from]
		},
		{
			'messageId':[MESSAGE_HASH],
			'senderId': [IPFS_PeerID],
			'sentAt': [UNIX_TIME_STAMP],
			'receivedAt': [UNIX_TIME_STAMP],
			'expiration': [UNIX_TIME_STAMP],
			'payload': [decrypted_payload_from]
		},
		{
			'messageId':[MESSAGE_HASH],
			'senderId': [IPFS_PeerID],
			'sentAt': [UNIX_TIME_STAMP],
			'receivedAt': [UNIX_TIME_STAMP],
			'expiration': [UNIX_TIME_STAMP],
			'payload': [decrypted_payload_from]
		}
	]
}
```

Messages contain the `messageId`, `senderId`, `sentAt`, `receivedAt`, `expiration`, and the decrypted `payload`. Any message may be removed by the recipient. This only removes the messages locally. Messages may also be automatically removed by inspecting the expiration.

