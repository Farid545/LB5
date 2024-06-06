package datastore

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	outFileName  = "current-data"
	segmentSize  = 10 * 1024 * 1024 // 10MB
	DeleteMarker = "DELETE"
)

var ErrNotFound = fmt.Errorf("record does not exist")

type hashIndex map[string]int64

type Db struct {
	out       *os.File
	outPath   string
	outOffset int64

	index   hashIndex
	mu      sync.RWMutex
	mergeMu sync.Mutex
}

// NewDb creates a new database
func NewDb(dir string) (*Db, error) {
	outputPath := filepath.Join(dir, outFileName)
	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	db := &Db{
		outPath: outputPath,
		out:     f,
		index:   make(hashIndex),
	}
	err = db.recover() // Recover data from the file
	if err != nil && err != io.EOF {
		return nil, err
	}
	return db, nil
}

const bufSize = 8192

// recover loads the index from the data file
func (db *Db) recover() error {
	fmt.Println("Recovering database...")
	input, err := os.Open(db.outPath)
	if err != nil {
		return err
	}
	defer input.Close()

	var buf [bufSize]byte
	in := bufio.NewReaderSize(input, bufSize)
	for err == nil {
		var (
			header, data []byte
			n            int
		)
		header, err = in.Peek(bufSize)
		if err == io.EOF {
			if len(header) == 0 {
				fmt.Println("Database recovered.")
				return err
			}
		} else if err != nil {
			return err
		}
		size := binary.LittleEndian.Uint32(header)

		if size < bufSize {
			data = buf[:size]
		} else {
			data = make([]byte, size)
		}
		n, err = in.Read(data)

		if err == nil {
			if n != int(size) {
				return fmt.Errorf("corrupted file")
			}

			var e entry
			e.Decode(data)
			if e.value == DeleteMarker {
				delete(db.index, e.key)
			} else {
				db.index[e.key] = db.outOffset
			}
			db.outOffset += int64(n)
		}
	}
	fmt.Println("Database recovered.")
	return err
}

func (db *Db) Close() error {
	fmt.Println("Closing database...")
	err := db.out.Close()
	if err != nil {
		return err
	}
	fmt.Println("Database closed.")
	return nil
}

func (db *Db) Get(key string) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	position, ok := db.index[key]
	if !ok {
		return "", ErrNotFound
	}

	file, err := os.Open(db.outPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Seek(position, 0)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(file)
	value, err := readValue(reader)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (db *Db) Put(key, value string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	e := entry{
		key:   key,
		value: value,
	}
	n, err := db.out.Write(e.Encode())
	if err == nil {
		db.index[key] = db.outOffset
		db.outOffset += int64(n)
	}

	if db.outOffset >= segmentSize {
		go db.mergeSegments()
	}

	return err
}

func (db *Db) Delete(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	e := entry{
		key:   key,
		value: DeleteMarker,
	}
	n, err := db.out.Write(e.Encode())
	if err == nil {
		delete(db.index, key)
		db.outOffset += int64(n)
	}

	if db.outOffset >= segmentSize {
		go db.mergeSegments()
	}

	return err
}

// mergeSegments merges data segments to save space
func (db *Db) mergeSegments() {
	fmt.Println("Starting segment merge")
	db.mergeMu.Lock()
	fmt.Println("Lock acquired for segment merge")
	defer db.mergeMu.Unlock()

	tempPath := db.outPath + ".temp"
	tempFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		fmt.Println("Failed to open temp file for merging:", err)
		return
	}
	defer tempFile.Close()

	db.mu.RLock()
	fmt.Println("Index keys iteration started")
	defer db.mu.RUnlock()

	for key, offset := range db.index {
		fmt.Printf("Processing key: %s at offset: %d\n", key, offset)
		file, err := os.Open(db.outPath)
		if err != nil {
			fmt.Println("Failed to open current out file for reading:", err)
			continue
		}
		_, err = file.Seek(offset, io.SeekStart)
		if err != nil {
			fmt.Println("Failed to seek in current out file:", err)
			file.Close()
			continue
		}

		reader := bufio.NewReader(file)
		value, err := readValue(reader)
		file.Close()
		if err != nil {
			fmt.Println("Failed to read value:", err)
			continue
		}

		e := entry{
			key:   key,
			value: value,
		}
		_, err = tempFile.Write(e.Encode())
		if err != nil {
			fmt.Println("Failed to write entry to temp file:", err)
			return
		}
		fmt.Println("Entry written to temp file successfully")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	err = db.out.Close()
	if err != nil {
		fmt.Println("Failed to close current out file:", err)
		return
	}
	fmt.Println("Current out file closed")

	err = os.Rename(tempPath, db.outPath)
	if err != nil {
		fmt.Println("Failed to rename temp file to current out file:", err)
		return
	}
	fmt.Println("Temp file renamed to current out file")

	db.out, err = os.OpenFile(db.outPath, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		fmt.Println("Failed to reopen current out file:", err)
		return
	}
	fmt.Println("Current out file reopened")
}

// readValue reads the value of an entry
func readValue(in *bufio.Reader) (string, error) {
	header, err := in.Peek(8)
	if err != nil {
		return "", err
	}
	keySize := int(binary.LittleEndian.Uint32(header[4:]))
	_, err = in.Discard(keySize + 8)
	if err != nil {
		return "", err
	}

	header, err = in.Peek(4)
	if err != nil {
		return "", err
	}
	valSize := int(binary.LittleEndian.Uint32(header))
	_, err = in.Discard(4)
	if err != nil {
		return "", err
	}

	data := make([]byte, valSize)
	n, err := in.Read(data)
	if err != nil {
		return "", err
	}
	if n != valSize {
		return "", fmt.Errorf("can't read value bytes (read %d, expected %d)", n, valSize)
	}

	return string(data), nil
}
