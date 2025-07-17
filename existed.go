//go:build js

package indexeddb

import (
	"context"

	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/hack-pad/safejs"
)

func (b *IndexeddbBackend) IsExisted(ctx context.Context, eventID string) bool {
	tx, err := b.db.Transaction(idb.TransactionReadOnly, storeNameEvents)
	if err != nil {
		logErr(err)
		return false
	}
	defer tx.Await(ctx)
	store, err := tx.ObjectStore(storeNameEvents)
	if err != nil {
		logErr(err)
		return false
	}
	rawID, err := safejs.ValueOf(eventID)
	if err != nil {
		logErr(err)
		return false
	}
	req, err := store.Get(rawID)
	if err != nil {
		logErr(err)
		return false
	}
	evt, err := req.Await(ctx)
	if err != nil {
		logErr(err)
		return false
	}
	if !evt.IsNull() && !evt.IsUndefined() {
		return true
	}
	return false
}
