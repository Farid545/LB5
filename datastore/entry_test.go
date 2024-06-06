package datastore

import (
	"testing"
)

func TestEntryEncoding(t *testing.T) {
	e := entry{
		key:   "key1",
		value: "value1",
	}
	encoded := e.Encode()
	decoded := entry{}
	decoded.Decode(encoded)
	if decoded.key != e.key {
		t.Fatalf("expected key %s, got %s", e.key, decoded.key)
	}
	if decoded.value != e.value {
		t.Fatalf("expected value %s, got %s", e.value, decoded.value)
	}
}
