//go:build js

package indexeddb

import (
	"context"
	"fmt"
	"runtime"
	"syscall/js"

	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/fiatjaf/eventstore"
	"github.com/hack-pad/safejs"
)

var _ eventstore.Store = (*IndexeddbBackend)(nil)

type IndexeddbBackend struct {
	db *idb.Database
}

func (b *IndexeddbBackend) Init() error {
	ctx := context.Background()
	var err error
	req, err := idb.Global().Open(ctx, databaseName, databaseVersion, upgrade)
	if err != nil {
		return err
	}
	b.db, err = req.Await(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (b *IndexeddbBackend) Close() {
	b.db.Close()
}

func (b *IndexeddbBackend) Reset() error {
	ctx := context.Background()
	req, err := idb.Global().DeleteDatabase(databaseName)
	if err != nil {
		return err
	}
	if err := req.Await(ctx); err != nil {
		return err
	}
	req2, err := idb.Global().Open(ctx, databaseName, databaseVersion, upgrade)
	if err != nil {
		return err
	}
	b.db, err = req2.Await(ctx)
	if err != nil {
		return err
	}
	return nil
}

// upgrade deletes the store if we have a new version. no migration for simplicity.
func upgrade(db *idb.Database, oldVersion, newVersion uint) error {
	if oldVersion < newVersion {
		names, err := db.ObjectStoreNames()
		if err != nil {
			return err
		}
		found := false
		for _, n := range names {
			if n == storeNameEvents {
				found = true
			}
		}

		if found {
			if err := db.DeleteObjectStore(storeNameEvents); err != nil {
				return err
			}
		}

		store, err := db.CreateObjectStore(storeNameEvents, idb.ObjectStoreOptions{AutoIncrement: false})
		if err != nil {
			return err
		}
		kpka, err := safejs.ValueOf([]any{keyKind, keyAuthor})
		if err != nil {
			return nil
		}
		if _, err := store.CreateIndex(
			idxKindAuthor,
			kpka,
			idb.IndexOptions{Unique: false, MultiEntry: false},
		); err != nil {
			return err
		}
		kpkm, err := safejs.ValueOf([]any{keyKind, keyMeta})
		if err != nil {
			return nil
		}
		if _, err := store.CreateIndex(
			idxKindMeta,
			kpkm,
			idb.IndexOptions{Unique: false, MultiEntry: false},
		); err != nil {
			return err
		}
		kpkta, err := safejs.ValueOf(keyKindTagAuthorArray)
		if err != nil {
			return nil
		}
		if _, err := store.CreateIndex(
			idxKindTagAuthor,
			kpkta,
			idb.IndexOptions{Unique: false, MultiEntry: true},
		); err != nil {
			return err
		}
	}
	return nil
}

func logf(format string, message ...any) {
	js.Global().Get("console").Call("log", js.ValueOf(fmt.Sprintf(format, message...)))
}

func logErr(err error) {
	_, f, l, _ := runtime.Caller(0)
	js.Global().Get("console").Call("error", js.ValueOf(fmt.Sprintf("%s %d: %s", f, l, err)))
}
