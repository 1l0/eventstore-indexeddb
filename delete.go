//go:build js

package indexeddb

import (
	"context"
	"syscall/js"

	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/hack-pad/safejs"
	"github.com/nbd-wtf/go-nostr"
)

func (b *IndexeddbBackend) DeleteEvent(ctx context.Context, evt *nostr.Event) error {
	tx, err := b.db.Transaction(idb.TransactionReadWrite, storeNameEvents)
	if err != nil {
		return err
	}
	store, err := tx.ObjectStore(storeNameEvents)
	if err != nil {
		return err
	}
	req, err := store.Delete(safejs.Safe(js.ValueOf(evt.ID)))
	if err != nil {
		return err
	}
	if err := req.Await(ctx); err != nil {
		return err
	}

	if err := tx.Await(ctx); err != nil {
		return err
	}
	return nil
}
