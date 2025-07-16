# eventstore-indexeddb

> [!CAUTION]
> This is not a general purpose eventstore; it is a small subset designed for a specific Nostr client.

Nostr eventstore for Wasm/Browser using indexeddb for the storage backend.

- meta only
  - relay list, user profile, relay info, group meta, etc
  -  we use `kind 2` for the relay info, not for the recommended server
- prefix search (subset of NIP-50) by name or URL for the meta
