package server

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/mauricio/gocached/store"
	"github.com/rainycape/memcache"
)

var serverPort = int32(10000)

func withServer(t *testing.T, entry func(store.Storage, Server, *memcache.Client)) {
	address := "localhost"
	items := store.New()
	currentPort := atomic.AddInt32(&serverPort, 1)
	server := New(currentPort, address, items)
	err := server.Start()

	if err != nil {
		t.Error("Failed to start server")
		return
	}

	defer server.Stop()

	client, err := memcache.New(fmt.Sprintf("%s:%d", address, currentPort))

	if err != nil {
		t.Error("Failed to connect to memcached server")
		return
	}

	defer client.Close()

	entry(items, server, client)
}

func TestSet(t *testing.T) {
	key := "foo"
	value := []byte("my value")

	withServer(t, func(items store.Storage, server Server, client *memcache.Client) {
		err := client.Set(&memcache.Item{Key: key, Value: value})

		if err != nil {
			t.Error("Set operation failed for client", err)
		}

		body, found := items.Get(key)

		if found {
			if !reflect.DeepEqual(value, body) {
				t.Error("Byte array should have been the same as the one put inside")
			}
		} else {
			t.Error("The key should have been found inside the collection")
		}
	})
}

func TestGetNotExisting(t *testing.T) {
	key := "foo"

	withServer(t, func(items store.Storage, server Server, client *memcache.Client) {
		_, err := client.Get(key)
		if err != memcache.ErrCacheMiss {
			t.Error("Failed to execute get operation", err)
		}
	})
}

func TestSetAndGet(t *testing.T) {
	key := "foo"
	value := []byte("my value")

	withServer(t, func(items store.Storage, server Server, client *memcache.Client) {
		err := client.Set(&memcache.Item{Key: key, Value: value})

		if err != nil {
			t.Error("Set operation failed for client", err)
		}

		item, err := client.Get(key)

		if err != nil {
			t.Error("Failed to perform get operation on key")
		}

		if !reflect.DeepEqual(value, item.Value) {
			t.Error("Byte array should have been the same as the one put inside")
		}
	})
}

func TestDeleteExisting(t *testing.T) {
	key := "foo"
	value := []byte("my value")

	withServer(t, func(items store.Storage, server Server, client *memcache.Client) {
		err := client.Set(&memcache.Item{Key: key, Value: value})

		if err != nil {
			t.Error("Set operation failed for client", err)
		}

		err = client.Delete(key)

		if err != nil {
			t.Error("Failed to delete key from server")
		}

		_, err = client.Get(key)

		if err != memcache.ErrCacheMiss {
			t.Error("Item should have been deleted")
		}
	})
}

func TestDeleteNonExisting(t *testing.T) {
	key := "foo"

	withServer(t, func(items store.Storage, server Server, client *memcache.Client) {
		err := client.Delete(key)

		if err != nil {
			t.Error("Failed to delete key from server")
		}
	})
}

func TestSetTwice(t *testing.T) {
	key := "foo"
	value := []byte("my value")
	otherValue := []byte("my other value")

	withServer(t, func(items store.Storage, server Server, client *memcache.Client) {
		err := client.Set(&memcache.Item{Key: key, Value: value})

		if err != nil {
			t.Error("Set operation failed for client", err)
		}

		item, err := client.Get(key)

		if err != nil {
			t.Error("Failed to perform get operation on key")
		}

		if !reflect.DeepEqual(value, item.Value) {
			t.Error("Byte array should have been the same as the one put inside")
		}

		err = client.Set(&memcache.Item{Key: key, Value: otherValue})

		if err != nil {
			t.Error("Set operation failed for client", err)
		}

		otherItem, otherErr := client.Get(key)

		if otherErr != nil {
			t.Error("Failed to perform get operation on key")
		}

		if !reflect.DeepEqual(otherValue, otherItem.Value) {
			t.Error("Byte array should have been the same as the one put inside")
		}

	})
}
