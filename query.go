//go:build js

package indexeddb

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/hack-pad/safejs"
	"github.com/nbd-wtf/go-nostr"
)

func (b *IndexeddbBackend) QueryEvents(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
	ch := make(chan *nostr.Event)
	tx, err := b.db.Transaction(idb.TransactionReadOnly, storeNameEvents)
	if err != nil {
		if err := tx.Abort(); err != nil {
			logErr(err)
		}
		close(ch)
		return nil, err
	}
	store, err := tx.ObjectStore(storeNameEvents)
	if err != nil {
		if err := tx.Abort(); err != nil {
			logErr(err)
		}
		close(ch)
		return nil, err
	}
	if err := validateFilter(filter); err != nil {
		if err := tx.Abort(); err != nil {
			logErr(err)
		}
		close(ch)
		return nil, err
	}

	go func() {
		defer func() {
			if err := tx.Await(ctx); err != nil {
				logErr(err)
			}
			close(ch)
		}()

		if len(filter.IDs) > 0 {
			for _, id_ := range filter.IDs {
				id, err := safejs.ValueOf(id_)
				if err != nil {
					logErr(err)
					return
				}
				req, err := store.Get(id)
				if err != nil {
					logErr(err)
					return
				}
				rawEvt, err := req.Await(ctx)
				if rawEvt.IsUndefined() || rawEvt.IsNull() {
					return
				}
				evt, err := valueToEvent(id, rawEvt)
				if err != nil {
					logErr(err)
					return
				}
				ch <- evt
			}
			return
		}

		if filter.Search != "" {
			idx, err := store.Index(idxKindMeta)
			if err != nil {
				logErr(err)
				return
			}
			if slices.Contains(filter.Kinds, nostr.KindProfileMetadata) {
				search := strings.TrimSpace(filter.Search)
				lower, err := safejs.ValueOf([]any{nostr.KindProfileMetadata, search})
				if err != nil {
					logErr(err)
					return
				}
				upper, err := safejs.ValueOf([]any{nostr.KindProfileMetadata, search + "\uffff"})
				if err != nil {
					logErr(err)
					return
				}
				rb, err := idb.NewKeyRangeBound(lower, upper, false, false)
				if err != nil {
					logErr(err)
					return
				}
				req, err := idx.OpenCursorRange(rb, idb.CursorNext)
				if err != nil {
					logErr(err)
					return
				}
				if err := handleRequest(ctx, ch, req); err != nil {
					logErr(err)
					return
				}
				return
			} else if slices.Contains(filter.Kinds, nostr.KindRecommendServer) {
				search := strings.TrimSpace(filter.Search)
				if !strings.HasPrefix(search, "wss://") && !strings.HasPrefix(search, "ws://") {
					search = "wss://" + search
				}
				lower, err := safejs.ValueOf([]any{nostr.KindRecommendServer, search})
				if err != nil {
					logErr(err)
					return
				}
				upper, err := safejs.ValueOf([]any{nostr.KindRecommendServer, search + "\uffff"})
				if err != nil {
					logErr(err)
					return
				}
				rb, err := idb.NewKeyRangeBound(lower, upper, false, false)
				if err != nil {
					logErr(err)
					return
				}
				req, err := idx.OpenCursorRange(rb, idb.CursorNext)
				if err != nil {
					logErr(err)
					return
				}
				if err := handleRequest(ctx, ch, req); err != nil {
					logErr(err)
					return
				}
				return
			}
			logErr(fmt.Errorf("unsupported kinds for search: %v", filter.Kinds))
			return
		}

		if len(filter.Tags) > 0 {
			idx, err := store.Index(idxKindTagAuthor)
			if err != nil {
				logErr(err)
				return
			}
			for _, kind := range filter.Kinds {
				for tagSymbol, tags := range filter.Tags {
					for _, tag := range tags {
						if len(filter.Authors) < 1 {
							kt := strconv.Itoa(kind) + tagSymbol + tag
							lower, err := safejs.ValueOf([]any{kt})
							if err != nil {
								logErr(err)
								return
							}
							upper, err := safejs.ValueOf([]any{kt + "\uffff"})
							if err != nil {
								logErr(err)
								return
							}
							rb, err := idb.NewKeyRangeBound(lower, upper, false, false)
							if err != nil {
								logErr(err)
								return
							}
							req, err := idx.OpenCursorRange(rb, idb.CursorNext)
							if err != nil {
								logErr(err)
								return
							}
							if err := handleRequest(ctx, ch, req); err != nil {
								logErr(err)
								return
							}
						} else {
							for _, author := range filter.Authors {
								kta := strconv.Itoa(kind) + tagSymbol + tag + author
								only, err := safejs.ValueOf(kta)
								if err != nil {
									logErr(err)
									return
								}
								rb, err := idb.NewKeyRangeOnly(only)
								if err != nil {
									logErr(err)
									return
								}
								req, err := idx.OpenCursorRange(rb, idb.CursorNext)
								if err != nil {
									logErr(err)
									return
								}
								if err := handleRequest(ctx, ch, req); err != nil {
									logErr(err)
									return
								}
							}
						}
					}
				}
			}
			return
		}

		if len(filter.Authors) > 0 {
			idx, err := store.Index(idxKindAuthor)
			if err != nil {
				logErr(err)
				return
			}
			for _, kind := range filter.Kinds {
				for _, author := range filter.Authors {
					only, err := safejs.ValueOf([]any{kind, author})
					if err != nil {
						logErr(err)
						return
					}
					rb, err := idb.NewKeyRangeOnly(only)
					if err != nil {
						logErr(err)
						return
					}
					req, err := idx.OpenCursorRange(rb, idb.CursorNext)
					if err != nil {
						logErr(err)
						return
					}
					if err := handleRequest(ctx, ch, req); err != nil {
						logErr(err)
						return
					}
				}
			}
			return
		}

		if len(filter.Kinds) > 0 {
			idx, err := store.Index(idxKindAuthor)
			if err != nil {
				logErr(err)
				return
			}
			for _, kind := range filter.Kinds {

				lower, err := safejs.ValueOf([]any{kind})
				if err != nil {
					logErr(err)
					return
				}
				upper, err := safejs.ValueOf([]any{kind, "\uffff"})
				if err != nil {
					logErr(err)
					return
				}
				rb, err := idb.NewKeyRangeBound(lower, upper, false, false)
				if err != nil {
					logErr(err)
					return
				}
				req, err := idx.OpenCursorRange(rb, idb.CursorNext)
				if err != nil {
					logErr(err)
					return
				}
				if err := handleRequest(ctx, ch, req); err != nil {
					logErr(err)
					return
				}
			}
			return
		}

	}()
	return ch, nil
}

func handleRequest(ctx context.Context, ch chan<- *nostr.Event, req *idb.CursorWithValueRequest) error {
	return req.Iter(ctx, func(cursor *idb.CursorWithValue) error {
		id, err := cursor.PrimaryKey()
		if err != nil {
			return err
		}
		if id.IsUndefined() {
			return fmt.Errorf("primary key is undefined")
		}
		rawEvt, err := cursor.Value()
		if err != nil {
			return err
		}
		evt, err := valueToEvent(id, rawEvt)
		if err != nil {
			return err
		}
		ch <- evt
		return nil
	})
}

func valueToEvent(rawID, rawEvent safejs.Value) (*nostr.Event, error) {
	d, err := rawID.String()
	if err != nil {
		return nil, err
	}
	k_, err := rawEvent.Get(keyKind)
	if err != nil {
		return nil, err
	}
	k, err := k_.Int()
	if err != nil {
		return nil, err
	}
	a_, err := rawEvent.Get(keyAuthor)
	if err != nil {
		return nil, err
	}
	a, err := a_.String()
	if err != nil {
		return nil, err
	}
	c_, err := rawEvent.Get(keyContent)
	if err != nil {
		return nil, err
	}
	c, err := c_.String()
	if err != nil {
		return nil, err
	}
	ca_, err := rawEvent.Get(keyCreatedAt)
	if err != nil {
		return nil, err
	}
	ca, err := ca_.Int()
	if err != nil {
		return nil, err
	}
	s_, err := rawEvent.Get(keySignature)
	if err != nil {
		return nil, err
	}
	s, err := s_.String()
	if err != nil {
		return nil, err
	}
	t_, err := rawEvent.Get(keyTagArray)
	if err != nil {
		return nil, err
	}
	t, err := valueToTags(t_)
	if err != nil {
		return nil, err
	}
	return &nostr.Event{
		ID:        d,
		PubKey:    a,
		CreatedAt: nostr.Timestamp(int64(ca)),
		Kind:      k,
		Tags:      t,
		Content:   c,
		Sig:       s,
	}, nil
}

func valueToTags(rawTags safejs.Value) (nostr.Tags, error) {
	l, err := rawTags.Length()
	if err != nil {
		return nostr.Tags{}, err
	}
	tags := make([]nostr.Tag, 0, l)
	for i := 0; i < l; i++ {
		rawTag, err := rawTags.Index(i)
		if err != nil {
			return nostr.Tags{}, err
		}
		ll, err := rawTag.Length()
		if err != nil {
			return nostr.Tags{}, err
		}
		tag := make([]string, 0, ll)
		for j := 0; j < ll; j++ {
			t, err := rawTag.Index(j)
			if err != nil {
				return nostr.Tags{}, err
			}
			tstr, err := t.String()
			if err != nil {
				return nostr.Tags{}, err
			}
			tag = append(tag, tstr)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func validateFilter(filter nostr.Filter) error {
	if len(filter.IDs) > 0 {
		if len(filter.Kinds) > 0 || len(filter.Authors) > 0 || filter.Search != "" || len(filter.Tags) > 0 {
			return fmt.Errorf("when querying IDs, no other fields are allowed")
		}
	} else {
		if len(filter.Kinds) < 1 && len(filter.Authors) < 1 && filter.Search == "" && len(filter.Tags) < 1 {
			return fmt.Errorf("no fields")
		}
	}
	return nil
}
