# Handshake Core Spec Doc

## Summary

This is the spec document for handshake-core, the underlying structure and protocol specification for the cli and GUI apps as well as a proposal on how data is structured and stored.

## Introduction

Handshake is designed to be a decentralized p2p symmetric encrypted communications tool based on in-person initialization of communication so that all future transmissions rely symmetric key cryptography. This is primarily a design for out-of-band communication in which communicating parties are worried about potential compromises in asymmetric encryption methodology, CA poisoning, or relying on centralized service providers for communications technology.

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
- Bob generates an initial symmetric key, along with the connection info for Alice to connect to, this initial data a json blob encoded into a QR code. This includes "strategy" info on how to "mix generated keys" including salt, an offset integer, and mixing strategy.
- When Alice scans the code, the has all the necessary info to start the process of setting up a direct local network exchange with bob.
- Alice replies to Bob by encrypting a json payload that includes her her hashmap pubkeys, and 2 lists of randomly generated lookupHash mixins as well as her 32 byte keys mixers.
- Bob replies with an encrypted payload of his hashmap pubkeys, a list of randomly generated lookupHashes as well as his 32 byte key mixers.
- Bob and Alice mix their keys using the specified strategy and hashing the results with blake2b256 to generate a new set of keys for Alice's list and Bob's list. This allows for both parties to participate in random number generation, but the secrets never travel over the wire (not even on the local network) since the strategy for dividing, offsetting, and salting are transmitted only through the QR code JSON payload
- Bob posts his first message to handshake by creating a new encrypted message payload and writing it to IPFS. This returns an immutable IPFS hash. This IPFS hash endpoint URL is then client-side encrypted into a hashmap message that is posted to the hashmap endpoint corresponding to his pubkey that was shared with Alice.
- Alice does the same.
- Bob queries Alice's hashmap endpoint to get her message. 
- Alice does the same.
- Initialization is complete.

In handshake, keys never pass over the wire. Generated in the QR code (or some other out-of-band strategy that all parties trust) are:

- preshared key to be used by secretbox
- salt for use on hashes
- data set names (names used by participants)
- mixin strategy (mixin ordering for hash as well as offsets for the list)

`hash(salt|mixin[0]|...|mixin[n-1])`


The generation of preshared keys involves generating a set of random lookup hashes and random keys. The default behavior is to generate some large number (possibly 100k) lookup hashes and keys, but this should be configurable by Bob to generate fewer (to force a shorter conversation).

The hashlookup list and 256 bit key will be expressed like this in json:

```
{
	"zzyfZivgjn75HNA/": "lo0nbGY9gMyO5ooEtxFOFijsYKRUtjAt3+jqqylaEBM=",
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

- decryption key lookup - upon receiving a new message, the lookupHash is used to find the decryption key. if the key hasn't been used, the `lookupHash` is used to get the key from the key table. Two possible error states are that they lookupHash isn't found or, it is found, but in the consumed lookupKey table, which which give a warning that a key was used twice. When a key is used and added to the consumed key list, it's entry should be removed from the lookup table. As cited above, the structure should look as follows:

```
{
  "zzhZYQ6aCq5jTa95": "EC4fN3NjSdN5TlDrGWuF8y40UmJep0OpTQS6EgngyQU=",
  "zzn0uGow4UDKkiUz": "P3AqPsCgUkMTdMRR5Z0y89Y8+Gnmcr19BOILcvOYs6E=",
  "zzpfZLoZfOAxcw1s": "w79zeAOz3Yi1H6OPWTxGWUow+mNmufDvT3Kxu3glm8c=",
  "zzpz2cIuut0QQqrk": "aQswMsU8kgh7Z3PO0VexAaviCAkiNE/379+dGpsiOJs=",
  "zzuRD/3NVtuE17pO": "pbiekyDpVm+tUJHx9iUqvNWIW8jcIZ4i7pnRBZugupY=",
  "zzwr/y/4qHYl7FZJ": "pKrO/GD58YyapUb8hWwYb9Mz8ajItoo34/OTmPvNR9A=",
  "zzy/CG+FcmRFdyVJ": "zqr4LMqVqauYHZrfMQyL85zmcOwuHqD80vwPKhrr2bY=",
  "zzy36vs6pK+rt7NB": "9U6ILHq9QkkA15AsZqK5q4VIv8/voqz531NwmI2BB1o=",
  "zzyfZivgjn75HNA/": "lo0nbGY9gMyO5ooEtxFOFijsYKRUtjAt3+jqqylaEBM="
}

```

NOTE: Each chat participant uses their own lookup table to compose messages and references the other participant lookuphashes when decrypting messages

- chat log - The message history table is used to store decrypted messages locally on the device. This is a time-series based data feed that blends messages form the sender and recipient in a single view.

```
[
	{
		'id':[IPFS_HASH],
		'sender': [SENDER_ID],
		'sent': [UNIX_NANO_TIME_STAMP],
		'received': [UNIX_TIME_STAMP],
		'ttl': [UNIX_TIME_STAMP],
		'data': [decrypted_payload_from]
	},
	{
		'id':[IPFS_HASH],
		'sender': [SENDER_ID],
		'sent': [UNIX_NANO_TIME_STAMP],
		'received': [UNIX_TIME_STAMP],
		'ttl': [UNIX_TIME_STAMP],
		'data': [decrypted_payload_from]
	},
	{
		'id':[IPFS_HASH],
		'sender': [SENDER_ID],
		'sent': [UNIX_NANO_TIME_STAMP],
		'received': [UNIX_TIME_STAMP],
		'ttl': [UNIX_TIME_STAMP],
		'data': [decrypted_payload_from]
	}
]
```

Messages contain the `id`, `sender`, `sent`, `received`, `ttl`, and the decrypted `payload`. Any message may be removed by the recipient. This only removes the messages locally. Messages may also be automatically removed by inspecting the expiration.

- Bob initiatates new chat.
- This generates a QR Code
- Alice scans the QR Code
- They exchange setup info
- Bob and Alice can now chat using handshake

Though the setup seems simple, there are some important details that need to be covered here, namely in exploring identity and file structure.

## Primary and Duress data stores

Primary and duress data store are identiacal in structure but serve very different purposes. They are completely isolated from one another and use seperate encryption keys. The duress data store is used if Bob or Alice are forced to decrypt handshake. This has a dedicated duress hashmap endpoint for each chat. It also has a small number of lookup hash and keys to send such data plus additional info. The duress hashmap endpoints + primary will be both be in the primary data store while only duress related endpoints and keys would be in the duress store. This allow

## Encryption Keys

On setting up handshake, the device will randomly generate 2 - 256 bit keys.

- the primary key will be the key used to encrypted and decrypt all data locally in the app on the device. 
- the secondary key is used for the durress storage

The user will not interact with this key directly, instead the user will generate a passcode (most likely 16 digts) that will generate a hash that will be used for a secretbox storage of the key.

This way, the user can change the password for either the primary or secondary key and the only thing that needs to be re-encrypted is the key locker, not all the encrypted data.

This doesn't mean that the primary key couldn't be changed, but it does mean that such a change would require quite a bit more work, potentially. There would need to be a change, update, and rollback pattern for failed attempts. This would have to be a `stop the world` opreration in that no changes to any of the data should be allowed to happen while such a change happens.


## Explorations in chat construction

A chat is composed of a personal identity (the self) and others. This definition of both the self and others carries no characteristics between chats. Each chat is compartmentalized inside its chat hash directory and includes a config, chat logs, and lookup hash tables.


```
global/
	fetch.json
	config.json
	profiles/
		a7f7a7da.json.secretbox
		38f828fa.json.secretbox
chats/
	b145da14/
		a7f7a7da/
			config.json.secretbox
			chatlog.json.secretbox
			lookups/
				48181616.json.secretbox
				214552a6.json.secretbox
		38f828fa/
			config.json.secretbox
			chatlog.json.secretbox
			lookups/
				1515adad.json.secretbox  
				deadbeef.json.secretbox
	d4452a12/
		a7f7a7da/
			config.json.secretbox
			chatlog.json.secretbox
			lookups/
				18181151.json.secretbox
				314242a6.json.secretbox
		38f828fa/
			config.json.secretbox
			chatlog.json.secretbox
			lookups/                     
				deaddead.json.secretbox
				beefbeef.json.secretbox	
```

`global/fetch.json` is used for the background fetch jobs on ios and android. It lists endpoints to query as well as the last datestamp. If an endpoint has a newer date stamp than any that is listed, a local alert goes off to notify the user that there are new messages

```

[
	{
		"url": https://prototypehashmap.sh/2DrjgbFyssWsFRteC5HpnZy3dKTUujhoUifkFwqmbPHTo6n3MX,
		"datestamp": 1539149509513671000
	},
	{
		"url": https://prototypehashmap.sh/2DrjgbFyssWsD4teC5HpnZy3dKTUujhoUifkFwqmbPHTo6n1RS,
		"datestamp": 1539149509513681432
	},
]

```

NOTE: one important thing to consider with fetch is that it not only potentially leaks hashes you are watching, but passively querying these endpoints on any internet connection the phone is connected to could be dangerous

`global/config.json` is used for application level configurations. This file is unencrypted, so any settings must be contained by application code defaults. These settings should primary be used for less sophisticated access attempts, more sophisticated attempts would be able to be altered outside of the secretbox containers. This file should possibly be encrypted by device level encryption if possible such as touchID or other platform specific crypto.

```
{
	"login_attempts": 5,	 // min is 3 Max is 10
}
```

`global/profiles/{profile_id}.json.secretbox` holds profile specific settings related to that profile. This is a place for settings, but is also an easy way to find the key that decrypts a specific hashed file name. The file name will match hashed directory names. When a user types in the passcode to access the application, each secretbox profile is attempted for decryption until one is successful. 

```
{
	"id": "a7f7a7da",
	"type": "primary",
	"key": "8MOOwWunzqyMqsR/6ciVnqX04ZMA766o4dEeE0D9VKk=",
	"delegated": [
		{ 
			"id": "38f828fa",
			"type": "duress",
			"key": "Qqlkj5PtYFvFudN2C9UHu0XmVGRm5y1SPM5jz29DmV0="
		}
	]
}
```

In this profile, we are looking at a primary profile. It also has access to a delegated profile that is type "duress". The key for both the primary and duress are stored here. This way a user can easily change the profile passcode without affecting the rest of the encrypted data. A primary profile needs access to the duress account to ensure that file sizes are identical and date stamps match. To an outside observer both the primary and duress profiles must look identical, otherwise the duress account, which would be used less frequently, would obviously be caught as the "fake" account.

`chats/` - is a directory that holds all chat data. subfolders here are randomly generated ids for the chats themselves. Inside of a chat directory are a set of folders based on profiles. These profile hashes should match those in the profile file. At minimum every chat should have a primary and duress subfolder.

`chats/{chat_id}/{profile_id}/` - in the chat folders hold complete configurations for a chat, this includes `config`, `chatlog`, and `lookups`

`chats/{chat_id}/{profile_id}/config.json.secretbox` - the chat config inside of a profile defines the chat settings, relates the lookups to identities, and holds things such as the full ed25519 private key for the user as well as the public keys for the other users involved in the chat.

```
{
	"id": "314242a6",
	"strategy": {
		"message": {
			"latest": {
				"type": "hashmap",
				"nodes": [
					"https://prototype.hashmap.sh"
				],
				"sigMethod": "nacl-sign-ed25519",
				"privateKeys": [
					"GgYoJG4UgdgAf77qRISk0pzo/s+BIvujBA5pzul+ykIYygyCbn1XdJbd5Le1ER+3M+yqptnx6RO8U7NBcOstqw==",
					"wI96q0clxZZgJ9xWvAcCIDUgYa8CDNxdmK+ebEEeRHa5SnsNqcY6TJ1JX5ougDH2ZK4MgOg6nJ1eWXwivon6Hw=="
				]
			},
			"storage": {
				"type": "ipfs",
				"nodes": [],
			},
			"cipher": "nacl-secretbox-16000"
		}
	},
	"identities": [
		{
			"id": "18181151":
			"alias": "bass-boat",
			"endpoints": [
				"2DrjgbFyssWsFRteC5HpnZy3dKTUujhoUifkFwqmbPHTo6n3MX",
				"2DrjgbDjr5KVtw388VJTN46Dcnp9dSkM7msabo4JoxeW4aKU7n"
			]
		}
	]
	"settings": {
		"maxTTL": 500,
	}
}
```

`id` - is the identity tied to lookups and chat logs
`strategy` - outlines message, storage, and cipher requirements. This it set on the initial handshake and cannot be changed. If changes need to be made, a new handshake is initiated.
`identities` - outlines the id for other chat participants, the alias used for chat logs, and the endpoints used by hashmap.
`settings` - is used for chat settings overrides, such as setting a maxTTL for messages (this can only be used to make TTL more aggressive than specified by the sender, not less).

`chats/{chat_id}/{profile_id}/chatlog.json.secretbox` - this is a json object that holds the timeseries data for the chat history. This includes references to the IFPS hash each message came from, the lookup hash identity associated, and other metadata such as TTL information. This is used for the UI view as well as

```
[
	{
		'id':[IPFS_HASH],
		'sender': [SENDER_ID],
		'sent': [UNIX_NANO_TIME_STAMP],
		'received': [UNIX_TIME_STAMP],
		'ttl': [UNIX_TIME_STAMP],
		'data': [decrypted_payload_from]
	},
	{
		'id':[IPFS_HASH],
		'sender': [SENDER_ID],
		'sent': [UNIX_NANO_TIME_STAMP],
		'received': [UNIX_TIME_STAMP],
		'ttl': [UNIX_TIME_STAMP],
		'data': [decrypted_payload_from]
	},
	{
		'id':[IPFS_HASH],
		'sender': [SENDER_ID],
		'sent': [UNIX_NANO_TIME_STAMP],
		'received': [UNIX_TIME_STAMP],
		'ttl': [UNIX_TIME_STAMP],
		'data': [decrypted_payload_from]
	}
]

```

`chats/{chat_id}/{profile_id}/lookups/{identity}.json.secretbox` - this is the directory that holds lookup hashes and keys for each identity in the chat. At first, we will only support 2 person chat (i think), but more could be introduced in the future.

```
{
  "zzhZYQ6aCq5jTa95": "EC4fN3NjSdN5TlDrGWuF8y40UmJep0OpTQS6EgngyQU=",
  "zzn0uGow4UDKkiUz": "P3AqPsCgUkMTdMRR5Z0y89Y8+Gnmcr19BOILcvOYs6E=",
  "zzpfZLoZfOAxcw1s": "w79zeAOz3Yi1H6OPWTxGWUow+mNmufDvT3Kxu3glm8c=",
  "zzpz2cIuut0QQqrk": "aQswMsU8kgh7Z3PO0VexAaviCAkiNE/379+dGpsiOJs=",
  "zzuRD/3NVtuE17pO": "pbiekyDpVm+tUJHx9iUqvNWIW8jcIZ4i7pnRBZugupY=",
  "zzwr/y/4qHYl7FZJ": "pKrO/GD58YyapUb8hWwYb9Mz8ajItoo34/OTmPvNR9A=",
  "zzy/CG+FcmRFdyVJ": "zqr4LMqVqauYHZrfMQyL85zmcOwuHqD80vwPKhrr2bY=",
  "zzy36vs6pK+rt7NB": "9U6ILHq9QkkA15AsZqK5q4VIv8/voqz531NwmI2BB1o=",
  "zzyfZivgjn75HNA/": "lo0nbGY9gMyO5ooEtxFOFijsYKRUtjAt3+jqqylaEBM="
}
```

## Initialization

The Initialization process is straight forward. 

1) Bob initializes a new chat and selects a chat strategy
2) Alice chooses to join
3) Bob generates some connection logic to share with Alice
4) Alice accepts this connection logic
5) The two devices share information
6) Chat begins

This section will attempt to outline this process in more detail.

## Message Handling

Assuming default Settings that Bob is using a chat strategy this includes:

- Hashmap for latest messsages
- IPFS for message storage
- secretbox for the cipher

Posting a message follows this process:

- Open Handshake App
- Authenticate with passcode, which decryptes chat data
- Enter a chat session
- Compose a message
- a temp file is generated for the submitted message during the submission process with status
- The message is client side encrypted with a randomly seleted lookuphash
- The lookuphash + key are deleted from the hashlist
- The message is submitted to the message storage and an IPFS hash is returned
- The IPFS hash is encrtypted with a randomly selected lookup hash
- The lookuphash + key are deleted from the hashlist
- The encrypted IPFS hash is submitted to hashmap using one or more hashmap private keys
- The message is added to the chat log
- the temp file for the message is deleted

Checking messages:

To check for messages from chat participants with a manual refresh (this is similar to how background queries work mechanically)

- Open a chat session
- trigger a refresh
- the chat config reads the identities and queries the hashmap endpoints outlined for that identity
- if a response is returned and the payload is valid, create a temp file for the update
- the if the lookup hash exists, the message is decrypted.
- Query the IPFS hash and if the hash responds, match the lookup hash to decrypt the payload
- if the lookup hash exists, attempt to decrypt the message
- if the message decrypts properly, update the chat log.
- if the decrypted IPFS hash references a parent hash, if the hash doesn't match a hash in the chat log, query the hash
- repeat the lookup and decrypt process recursively until either an IPFS hash is matched or a lookuphash cannot be found

## House keeping

When the app is open, it should scan through all chats looking to clean up chat logs. This should happen when the app is open, but also as a part of background operations on some frequency. This also goes for querying identity hashmap endpoints for new messages.