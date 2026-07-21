package provider

import (
	"bytes"
	"encoding/json"
)

// Canonicalize re-encodes a JSON document with sorted object keys, compact
// whitespace, and preserved number text, so envelopes from different
// transports can be compared byte-for-byte.
func Canonicalize(raw []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}
