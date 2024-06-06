package datastore

import (
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"
	"log"
)

// TestDb_Put tests the Put and Get methods of the database
func TestDb_Put(t *testing.T) {
	// Create a temporary directory for the test
	dir, err := ioutil.TempDir("", "test-db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a new database
	db, err := NewDb(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	pairs := [][]string{
		{"key1", "value1"},
		{"key2", "value2"},
		{"key3", "value3"},
	}

	// Test putting and getting key-value pairs
	t.Run("put/get", func(t *testing.T) {
		for _, pair := range pairs {
			log.Printf("Putting key: %s value: %s", pair[0], pair[1])
			err := db.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("Cannot put %s: %s", pair[0], err)
			}
			value, err := db.Get(pair[0])
			if err != nil {
				t.Errorf("Cannot get %s: %s", pair[0], err)
			}
			if value != pair[1] {
				t.Errorf("Bad value returned expected %s, got %s", pair[1], value)
			}
		}
	})

	// Test if the file grows with many entries
	t.Run("file growth", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			err := db.Put("key"+strconv.Itoa(i), "value"+strconv.Itoa(i))
			if err != nil {
				t.Errorf("Cannot put key%d: %s", i, err)
			}
		}
	})
	
	// Test if a new database process recovers the data
	t.Run("new db process", func(t *testing.T) {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		db, err = NewDb(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, pair := range pairs {
			value, err := db.Get(pair[0])
			if err != nil {
				t.Errorf("Cannot get %s: %s", pair[0], err)
			}
			if value != pair[1] {
				t.Errorf("Bad value returned expected %s, got %s", pair[1], value)
			}
		}
	})
}

// TestSegmentMerge tests the segment merging functionality
func TestSegmentMerge(t *testing.T) {
	// Create a temporary directory for the test
	dir, err := ioutil.TempDir("", "test-db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	db, err := NewDb(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var wg sync.WaitGroup

	pairs := [][]string{
		{"key1", "value1"},
		{"key2", "value2"},
		{"key3", "value3"},
	}

	for _, pair := range pairs {
		err := db.Put(pair[0], pair[1])
		if err != nil {
			t.Errorf("Cannot put %s: %s", pair[0], err)
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.mergeSegments()
	}()

	wg.Wait()

	for _, pair := range pairs {
		value, err := db.Get(pair[0])
		if err != nil {
			t.Errorf("Cannot get %s after merge: %s", pair[0], err)
		}
		if value != pair[1] {
			t.Errorf("Bad value returned after merge expected %s, got %s", pair[1], value)
		}
	}
}
