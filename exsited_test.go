//go:build js

package indexeddb

import (
	"fmt"
	"testing"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/sdk"
)

func TestExisted(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatal(err)
	}
	id, _, err := db.saveProfile(sdk.ProfileMetadata{
		Name:        "jack",
		DisplayName: "Jack",
		About:       "sup",
	})
	if err != nil {
		t.Fatal(err)
	}
	if exsited := db.IsExisted(db.ctx, id); !exsited {
		t.Fatal(fmt.Errorf("expected: true, actual: %t", exsited))
	}
	random := nostr.Generate()
	if exsited := db.IsExisted(db.ctx, random.Hex()); exsited {
		t.Fatal(fmt.Errorf("expected: false, actual: %t", exsited))
	}
}
