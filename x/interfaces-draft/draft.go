/*
	The purpose of this code is to describe the basic interfaces,
	methods, and funcationality for handshake core. This is subject
	to change and will change greatly as ideas solidify
*/
package handshake

// ideas for structs

// Session
type Session struct {
	Profiles map[string]Profile
	Settings GlobalSettings
}

// represents a profile that has been accessed
// this would contain successfully decrypted profile data
type Profile struct {
	Settings ProfileSettings
}

type ProfileSettings struct {
	//
}

type Chat struct {
	Settings ChatSettings
}

type ChatSettings struct {
	//
}

type Strategy struct {
}

type ChatLog struct {
}

type Lookup struct {
}

type GlobalSettings struct {
	//
}
