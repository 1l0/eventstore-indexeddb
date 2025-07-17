//go:build js

package indexeddb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/hack-pad/safejs"
	"github.com/nbd-wtf/go-nostr"
)

func (b *IndexeddbBackend) SaveEvent(ctx context.Context, evt *nostr.Event) error {
	if evt == nil {
		return nil
	}

	// validate kinds
	if !nostr.IsReplaceableKind(evt.Kind) && evt.Kind != nostr.KindRecommendServer && !nostr.IsAddressableKind(evt.Kind) {
		return nil
	}

	tx, err := b.db.Transaction(idb.TransactionReadWrite, storeNameEvents)
	if err != nil {
		return err
	}
	store, err := tx.ObjectStore(storeNameEvents)
	if err != nil {
		return err
	}

	meta, err := ParseMeta(evt)
	if err != nil {
		return err
	}
	var metaValue any = nil
	if meta.Name != "" {
		metaValue = meta.Name
	} else if meta.URL != "" {
		metaValue = meta.URL
	}

	k := strconv.Itoa(evt.Kind)
	p := evt.PubKey
	tags := []any{}
	kta := []any{}
	addressable := nostr.IsAddressableKind(evt.Kind)

	for _, tag := range evt.Tags {
		if len(tag) < 2 || len(tag[1]) < 1 {
			continue
		}

		tagjs := []any{}
		for _, t := range tag {
			tagjs = append(tagjs, t)
		}
		tags = append(tags, tagjs)

		if len(tag[0]) != 1 {
			continue
		}

		if addressable && tag[0] == "d" {
			kta = append(kta, k+tag[0]+tag[1]+p)
		}
	}

	obj := map[string]any{
		keyKind:               evt.Kind,
		keyAuthor:             evt.PubKey,
		keyContent:            evt.Content,
		keyTagArray:           tags,
		keyCreatedAt:          int64(evt.CreatedAt),
		keySignature:          evt.Sig,
		keyKindTagAuthorArray: kta,
		keyMeta:               metaValue,
	}

	rawID, err := safejs.ValueOf(evt.ID)
	if err != nil {
		return err
	}
	rawObj, err := safejs.ValueOf(obj)
	if err != nil {
		return err
	}

	_, err = store.PutKey(rawID, rawObj)

	if err != nil {
		return err
	}
	if err := tx.Await(ctx); err != nil {
		return err
	}
	return nil
}

type Meta struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

func ParseMeta(event *nostr.Event) (meta Meta, err error) {
	if event == nil {
		return Meta{}, fmt.Errorf("event is nil")
	}
	if event.Kind != nostr.KindProfileMetadata &&
		event.Kind != nostr.KindRecommendServer {
		return Meta{}, nil
	}
	if er := json.Unmarshal([]byte(event.Content), &meta); er != nil {
		err = er
	}
	if meta.Name != "" {
		meta.Name = strings.ToLower(strings.TrimSpace(meta.Name))
	}
	if meta.URL != "" {
		meta.URL = nostr.NormalizeURL(meta.URL)
	}
	return meta, err
}
