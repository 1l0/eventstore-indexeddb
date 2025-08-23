//go:build js

package indexeddb

import (
	"fiatjaf.com/nostr"
)

func (b *IndexeddbBackend) CountEvents(nostr.Filter) (uint32, error) {
	// TODO:
	return 0, nil
}
