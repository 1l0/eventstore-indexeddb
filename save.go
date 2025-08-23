//go:build js

package indexeddb

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"fiatjaf.com/nostr"
	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/hack-pad/safejs"
)

func (b *IndexeddbBackend) SaveEvent(evt nostr.Event) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// validate kinds
	if !evt.Kind.IsReplaceable() && evt.Kind != nostr.KindRecommendServer && !evt.Kind.IsAddressable() {
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

	k := strconv.Itoa(int(evt.Kind))
	p := evt.PubKey.Hex()
	tags := []any{}
	kta := []any{}
	addressable := evt.Kind.IsAddressable()

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

	sig := hex.EncodeToString(evt.Sig[:])

	obj := map[string]any{
		keyKind:               evt.Kind.Num(),
		keyAuthor:             evt.PubKey.Hex(),
		keyContent:            evt.Content,
		keyTagArray:           tags,
		keyCreatedAt:          int64(evt.CreatedAt),
		keySignature:          sig,
		keyKindTagAuthorArray: kta,
		keyMeta:               metaValue,
	}

	rawID, err := safejs.ValueOf(evt.ID.Hex())
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

func ParseMeta(event nostr.Event) (meta Meta, err error) {
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
