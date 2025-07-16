//go:build js

package indexeddb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/fiatjaf/eventstore"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/sdk"
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
	sk := nostr.GeneratePrivateKey()

	p, err := json.Marshal(profile)
	if err != nil {
		return "", "", err
	}
	evt := &nostr.Event{
		Kind:      nostr.KindProfileMetadata,
		Content:   string(p),
		CreatedAt: nostr.Now(),
	}
	if err := evt.Sign(sk); err != nil {
		return "", "", err
	}
	if err := db.SaveEvent(db.ctx, evt); err != nil && err != eventstore.ErrDupEvent {
		return "", "", err
	}
	return evt.ID, evt.PubKey, nil
}

func (db *DB) saveGroupMeta(id, name string) (string, string, error) {
	sk := nostr.GeneratePrivateKey()

	evt := &nostr.Event{
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
		logf("fail on sign")
		return "", "", err
	}
	if err := db.SaveEvent(db.ctx, evt); err != nil && err != eventstore.ErrDupEvent {
		logf("fail on save")
		return "", "", err
	}
	return evt.ID, evt.PubKey, nil
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
		IDs: []string{id},
	}
	ch, err := db.QueryEvents(db.ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for evt := range ch {
		count++
		if evt.ID != id {
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
		Kinds:  []int{0},
		Search: "jack",
	}
	ch, err := db.QueryEvents(db.ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for evt := range ch {
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
		Kinds:  []int{0},
		Search: "bob",
	}
	ch, err = db.QueryEvents(db.ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	for evt := range ch {
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
		Kinds:   []int{nostr.KindSimpleGroupMetadata},
		Authors: []string{pk},
		Tags:    nostr.TagMap{"d": []string{id}},
	}
	ch, err := db.QueryEvents(db.ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for range ch {
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
		Kinds:   []int{0},
		Authors: []string{pk, pk2},
	}
	ch, err := db.QueryEvents(db.ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for range ch {
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
		Kinds: []int{0},
	}
	ch, err := db.QueryEvents(db.ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for range ch {
		count++
	}
	if count != 2 {
		t.Fatal(fmt.Errorf("count expect 2, actual: %d", count))
	}
	filter = nostr.Filter{
		Kinds: []int{1},
	}
	ch, err = db.QueryEvents(db.ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		count++
	}
	if count != 2 {
		t.Fatal(fmt.Errorf("unexpected kind response"))
	}
}
