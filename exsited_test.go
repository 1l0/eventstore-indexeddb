package indexeddb

import (
	"fmt"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/sdk"
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
	random := nostr.GeneratePrivateKey()
	if exsited := db.IsExisted(db.ctx, random); exsited {
		t.Fatal(fmt.Errorf("expected: false, actual: %t", exsited))
	}
}
