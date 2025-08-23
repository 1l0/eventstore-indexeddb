//go:build js

package indexeddb

import (
	"context"
	"syscall/js"

	"fiatjaf.com/nostr"
	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/hack-pad/safejs"
)

func (b *IndexeddbBackend) DeleteEvent(id nostr.ID) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := b.db.Transaction(idb.TransactionReadWrite, storeNameEvents)
	if err != nil {
		return err
	}
	store, err := tx.ObjectStore(storeNameEvents)
	if err != nil {
		return err
	}
	req, err := store.Delete(safejs.Safe(js.ValueOf(id.Hex())))
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
