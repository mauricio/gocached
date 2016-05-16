package store

import (
	"reflect"
	"testing"
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

func TestDeleteExisting(t *testing.T) {
	value := []byte{1, 2, 3}
	key := "some key"
	items := New()

	items.Put(key, value)
	items.Delete(key)

	_, found := items.Get(key)

	if found {
		t.Error("Should be false meaning the item is not there anymore")
	}
}

func TestDeleteNotExisting(t *testing.T) {
	key := "some key"
	items := New()

	items.Delete(key)
	_, found := items.Get(key)

	if found {
		t.Error("Should be false meaning the item is not there at all")
	}
}
