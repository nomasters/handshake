# Handshake Core Spec Doc

## Summary

This is the spec document for handshake-core, the underlying structure and protocol specification for the cli and GUI apps as well as a proposal on how data is structured and stored.

## Introduction

Handshake is designed to be a decentralized p2p encrypted communications tool based on in-person initialization of communication so that all future transmissions rely on symmetric key cryptography. This is primarily a design for out-of-band communication in which communicating parties are worried about potential compromises in asymmetric encryption methodology, CA poisoning, or relying on centralized service providers for communications technology.

Handshake is designed initially to work on IPFS and hashmap, but there are no technical reasons other backends couldn't be supported. In fact, in discussions, the core developers want to encourage a sort of "strategies" approach in which handshake participants might pick and choose tooling that works best for their needs. For the sake of focus, this specification utilizes hashmap and IPFS, but tooling like Iota MAM, ethereum smart contracts, and other systems should be able to be incorporated in the future.

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

- When a new chat is initialized, either as an initiator or a joiner, a set of keys and config files are generated specifically for the chat session. No unique info is shared across chats. There is no concept of a "user identity" that persists since both parties must meet in person and this tool is designed to be low-knowledge out-of-band p2p encrypted chat tool.
- In its default configuration, the chat tool uses IPFS public gateways to submit data to the IPFS network and it uses a hashmap gateway to submit a client side encrypted message that references the "latest message" IPFS hash.
- Bob generates an initial 32 byte symmetric key, along with the port and route for Alice to connect to, this initial data a json blob encoded into a QR code. This includes "strategy" info on how to "mix generated keys" including salt, an offset integer, and mixing strategy.
- When Alice scans the code, the has all the necessary info to start the process of setting up a direc local network exchange with bob.
- Alice replies to Bob by encrypting a json payload that includes her her hashmap pubkeys, and list of randomly generated lookupHashes as well as 32 byte key mixers.
- Bob replies with an encrypted payload of his hashmap pubkeys, a list of randomly generated lookupHashes as well as his 32 byte key mixers.
- Bob and Alice mix their keys using the specified strategy and hashing the results with blake2b256 to generate a new set of keys for Alice's list and Bob's list. This allows for both parties to participate in random number generation, but the secrets never travel over the wire (not even on the local network) since the strategy for dividing, offsetting, and salting are transmitted only through the QR code JSON payload
- Bob posts his first message to handshake by creating a new encrypted message payload and writing it to IPFS. This returns an immutable IPFS hash. This IPFS hash endpoint URL is then client-side encrypted into a hashmap message that is posted to the hashmap endpoint corresponding to his pubkey that was shared with Alice.
- Alice does the same.
- Bob queries Alice's hashmap endpoint to get her message. 
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
    "method": "nacl-secretbox-16000",
    "lookup": "BASE_64_ENCODED_STRING",
    "data": "BASE_64_ENCODED_STRING"
}
```

It includes a description of the cryptography used. The initial implementation will use `secretbox` which requires a chunk size for larger messages to be processed efficiently. The the `hash` is the lookup hash that the message recipient would use to lookup which secret key to use. The data is encoded with base64 encoding and follows the `secretbox` spec in which the chunk nonce and authentication bytes exists before each encrypted chunk in the data.

Once decrypted, the data would look something like this:

```
{
    "parent": [IPFS_HASH_OF_PREVIOUS_MESSAGE],
    "timestamp": [UNIX_NANO_TIME_STAMP],
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
- `timestamp` is the unix_time in nano seconds, if no `timestamp` is present, the app will used received time. This is used to help weave two IPNS conversation endpoints together
- a message can contain media and a body,
- `media` is a place holder for future work, but will allow pictures and video to be included in a message.
- `message` is for the message body of the payload and must be utf-8.
- `ttl` is the TTL before the decrypted message is destroyed on the client. 

Upon receiving a message and successfully decrypting the message, the key is destroyed.

The `lookup` is the index for various types of lookups:

- decryption key lookup - upon receiveing a new message, the lookupHash is used to find the decryption key. if the key hasn't been used, the `lookupHash` is used to get the key from the key table. Two possible error states are that they lookupHash isn't found or, it is found, but in the consumed lookupKey table, which which give a warning that a key was used twice. When a key is used and added to the consumed key list, it's entry should be removed from the lookup table. As cited above, the structure should look as follows:

```
{
	"e1f1534d26d2": e5ac434796dab5f5164796d4"
	...
}

```

NOTE: Bob and Alice each have their own lookupHash tables, Bob's was generated by Bob and Alice by Alice on initialization

- consumed hash - A consumed key is linked to a specific lookuphash + IPFS hash. the lookupKey + hash gives an immutable consumption hash pair. If a lookup key is ever used twice, the ipfs hashes will differ and the lookup should be rejected (and the recipient notified). The data structure should look as follows:

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

