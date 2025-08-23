//go:build js

package indexeddb

import (
	"fmt"

	"fiatjaf.com/nostr"
	"github.com/fiatjaf/eventstore"
)

func (b *IndexeddbBackend) ReplaceEvent(evt nostr.Event) error {

	filter := nostr.Filter{Limit: 1, Kinds: []nostr.Kind{evt.Kind}, Authors: []nostr.PubKey{evt.PubKey}}
	if evt.Kind.IsAddressable() {
		filter.Tags = nostr.TagMap{"d": []string{evt.Tags.GetD()}}
	}

	itr := b.QueryEvents(filter, 1000)

	shouldStore := true
	for previous := range itr {
		if isOlder(previous, evt) {
			if err := b.DeleteEvent(previous.ID); err != nil {
				return fmt.Errorf("failed to delete event for replacing: %w", err)
			}
		} else {
			shouldStore = false
		}
	}

	if shouldStore {
		if err := b.SaveEvent(evt); err != nil && err != eventstore.ErrDupEvent {
			return fmt.Errorf("failed to save: %w", err)
		}
	}

	return nil
}

func isOlder(previous, next nostr.Event) bool {
	return previous.CreatedAt < next.CreatedAt ||
		(previous.CreatedAt == next.CreatedAt && previous.ID.Hex() > next.ID.Hex())
}
