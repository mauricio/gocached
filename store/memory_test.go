package store

import (
	"testing"
	"reflect"
)

func TestPut(t *testing.T) {
	value := []byte{1, 2, 3}
	key := "some key"
	items := New()

	items.Put(key, value)

	bytes, returned := items.Get(key)

	if !returned {
		t.Error("Should have found the value")
	}

	if !reflect.DeepEqual(value, bytes) {
		t.Error("Byte array should have been the same as the one put inside")
	}
}
