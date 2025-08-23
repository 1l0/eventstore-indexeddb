//go:build js

package indexeddb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/sdk"
)

type DB struct {
	*IndexeddbBackend
	ctx context.Context
}

func newDB() (*DB, error) {
	db := &IndexeddbBackend{}
	if err := db.Reset(); err != nil {
		return nil, err
	}

	ctx := context.Background()
	return &DB{
		IndexeddbBackend: db,
		ctx:              ctx,
	}, nil
}

func (db *DB) saveProfile(profile sdk.ProfileMetadata) (string, string, error) {
	sk := nostr.Generate()

	p, err := json.Marshal(profile)
	if err != nil {
		return "", "", err
	}
	evt := nostr.Event{
		Kind:      nostr.KindProfileMetadata,
		Content:   string(p),
		CreatedAt: nostr.Now(),
	}
	if err := evt.Sign(sk); err != nil {
		return "", "", err
	}
	if err := db.ReplaceEvent(evt); err != nil {
		return "", "", err
	}
	return evt.ID.Hex(), evt.PubKey.Hex(), nil
}

func (db *DB) saveGroupMeta(id, name string) (string, string, error) {
	sk := nostr.Generate()

	evt := nostr.Event{
		Kind:      nostr.KindSimpleGroupMetadata,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"d", id},
			nostr.Tag{"name", name},
			nostr.Tag{"public"},
			nostr.Tag{"open"},
		},
	}
	if err := evt.Sign(sk); err != nil {
		slog.Warn("fail on sign")
		return "", "", err
	}
	if err := db.ReplaceEvent(evt); err != nil {
		slog.Warn("fail on save")
		return "", "", err
	}
	return evt.ID.Hex(), evt.PubKey.Hex(), nil
}

func TestID(t *testing.T) {
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
	filter := nostr.Filter{
		IDs: []nostr.ID{nostr.MustIDFromHex(id)},
	}
	itr := db.QueryEvents(filter, 1000)
	count := 0
	for evt := range itr {
		count++
		if evt.ID.Hex() != id {
			t.Fatalf("id mismatch expeected: %s, actual: %s", id, evt.ID)
		}
	}
	if count != 1 {
		t.Fatalf("count expected: 1, actual: %d", count)
	}

}

func TestSearch(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.saveProfile(sdk.ProfileMetadata{
		Name:        "jack",
		DisplayName: "Jack",
		About:       "sup",
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.saveProfile(sdk.ProfileMetadata{
		Name:        "bob",
		DisplayName: "Bob",
		About:       "Yo",
	}); err != nil {
		t.Fatal(err)
	}
	filter := nostr.Filter{
		Kinds:  []nostr.Kind{0},
		Search: "jack",
	}
	itr := db.QueryEvents(filter, 1000)
	count := 0
	for evt := range itr {
		count++
		meta, err := ParseMeta(evt)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(meta.Name, "jack") {
			t.Errorf("meta.Name expect: jack*, actual: %s", meta.Name)
		}
	}
	if count == 0 {
		t.Fatal(fmt.Errorf("jack not found"))
	}
	if count > 1 {
		t.Fatal(fmt.Errorf("too many jacks"))
	}
	filter = nostr.Filter{
		Kinds:  []nostr.Kind{0},
		Search: "bob",
	}
	itr = db.QueryEvents(filter, 1000)
	for evt := range itr {
		count++
		meta, err := ParseMeta(evt)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(meta.Name, "bob") {
			t.Errorf("meta.Name expect: bob*, actual: %s", meta.Name)
		}
	}
	if count == 1 {
		t.Fatal(fmt.Errorf("bob not found"))
	}
	if count > 2 {
		t.Fatal(fmt.Errorf("too many bobs"))
	}
}

func TestKindTagAuthor(t *testing.T) {
	db, err := newDB()

	if err != nil {
		t.Fatal(err)
	}
	id := "asdf"
	name := "ASDF"
	_, pk, err := db.saveGroupMeta(id, name)
	if err != nil {
		t.Fatal(err)
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindSimpleGroupMetadata},
		Authors: []nostr.PubKey{nostr.MustPubKeyFromHex(pk)},
		Tags:    nostr.TagMap{"d": []string{id}},
	}
	itr := db.QueryEvents(filter, 1000)
	count := 0
	for range itr {
		count++
	}
	if count != 1 {
		t.Fatal(fmt.Errorf("count expect 1, actual: %d", count))
	}
}

func TestKindAuthor(t *testing.T) {
	db, err := newDB()

	if err != nil {
		t.Fatal(err)
	}
	_, pk, err := db.saveProfile(sdk.ProfileMetadata{
		Name:        "jack",
		DisplayName: "Jack",
		About:       "sup",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, pk2, err := db.saveProfile(sdk.ProfileMetadata{
		Name:        "bob",
		DisplayName: "Bob",
		About:       "Yo",
	})
	if err != nil {
		t.Fatal(err)
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{0},
		Authors: []nostr.PubKey{nostr.MustPubKeyFromHex(pk), nostr.MustPubKeyFromHex(pk2)},
	}
	itr := db.QueryEvents(filter, 1000)
	count := 0
	for range itr {
		count++
	}
	if count != 2 {
		t.Fatal(fmt.Errorf("count expect 2, actual: %d", count))
	}
}

func TestKind(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.saveProfile(sdk.ProfileMetadata{
		Name:        "jack",
		DisplayName: "Jack",
		About:       "sup",
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.saveProfile(sdk.ProfileMetadata{
		Name:        "bob",
		DisplayName: "Bob",
		About:       "Yo",
	}); err != nil {
		t.Fatal(err)
	}
	filter := nostr.Filter{
		Kinds: []nostr.Kind{0},
	}
	itr := db.QueryEvents(filter, 1000)
	count := 0
	for range itr {
		count++
	}
	if count != 2 {
		t.Fatal(fmt.Errorf("count expect 2, actual: %d", count))
	}
	filter = nostr.Filter{
		Kinds: []nostr.Kind{1},
	}
	itr = db.QueryEvents(filter, 1000)
	for range itr {
		count++
	}
	if count != 2 {
		t.Fatal(fmt.Errorf("unexpected kind response"))
	}
}
