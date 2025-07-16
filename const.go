//go:build js

package indexeddb

const (
	databaseVersion = 4
)

const (
	databaseName    = "eventstore"
	storeNameEvents = "events"

	keyKind               = "k"
	keyAuthor             = "a"
	keyContent            = "c"
	keyTagArray           = "t"
	keyCreatedAt          = "ca"
	keySignature          = "s"
	keyKindTagAuthorArray = "kta"
	keyMeta               = "m"

	idxKindAuthor    = "xka"
	idxKindMeta      = "xkm"
	idxKindTagAuthor = "xkta"
)
