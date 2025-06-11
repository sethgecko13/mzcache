package mzcache

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
)

const stdTestMessage = "result was incorrect, got: %v, want: %v."
const errTestMessage = "result was incorrect, got error: %v"

func testWrite() error {
	expected := "value1\nvalue2\nvalue3"
	key := "test1"
	err := Write(key, expected)
	if err != nil {
		return err
	}
	result, _ := Read(key, 1)
	if result != expected {
		return fmt.Errorf(stdTestMessage, result, expected)
	}
	return nil
}

func TestWrite(t *testing.T) {
	t.Parallel()
	err := testWrite()
	if err != nil {
		t.Errorf("got an unexpected error: %v", err.Error())
	}
}
func TestWriteEmptyString(t *testing.T) {
	t.Parallel()
	value := ""
	key := "empty"
	err := Write(key, value)
	if err != ErrCacheEmptyString {
		t.Errorf("got %v, expected: %v", err.Error(), ErrCacheEmptyString.Error())
	}
}
func TestReadExpired(t *testing.T) {
	t.Parallel()
	expected := "value1\nvalue2\nvalue3\nvalue4"
	key := "expired"
	err := Write(key, expected)
	if err != nil {
		t.Errorf(errTestMessage, err.Error())
	}
	_, err = Read(key, -1)
	var expiredErr *ErrCacheExpired
	if !errors.As(err, &expiredErr) {
		t.Errorf(errTestMessage, err.Error())
	}
}
func TestReadHit(t *testing.T) {
	t.Parallel()
	expected := "value1\nvalue2\nvalue3"
	key := "test2"
	err := Write(key, expected)
	if err != nil {
		t.Errorf(errTestMessage, err.Error())
	}
	result, err := Read(key, 1)
	if err != nil {
		t.Errorf(errTestMessage, err.Error())
	}
	if result != expected {
		t.Errorf(stdTestMessage, result, expected)
	}
}

func TestReadInvalidDirectory(t *testing.T) {
	// Don't run in parallel, will break other tests
	os.Setenv("MZ_CACHE_DIR", "/var/tmp/blah")
	defer func() {
		os.Setenv("MZ_CACHE_DIR", "/var/tmp/mzcache")
	}()
	_, err := Read("invalid_directory", 1)
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf(stdTestMessage, err, ErrCacheMiss.Error())
	}
}
func TestGetCacheDirEmpty(t *testing.T) {
	// Don't run in parallel, will break other tests
	newDir := ""
	os.Setenv("MZ_CACHE_DIR", newDir)
	defer func() {
		os.Setenv("MZ_CACHE_DIR", "/var/tmp/mzcache")
	}()
	result := getCacheDir()
	expected := "/var/tmp/mzcache"
	if result != expected {
		t.Errorf(stdTestMessage, result, expected)
	}
}
func TestGetCacheDirChanged(t *testing.T) {
	// Don't run in parallel, will break other tests
	newDir := "/var/tmp/blah"
	os.Setenv("MZ_CACHE_DIR", newDir)
	defer func() {
		os.Setenv("MZ_CACHE_DIR", "/var/tmp/mzcache")
	}()
	result := getCacheDir()
	if result != newDir {
		t.Errorf(stdTestMessage, result, newDir)
	}
}
func TestGetCacheDir(t *testing.T) {
	// Don't run in parallel, will break other tests
	result := getCacheDir()
	expected := "/var/tmp/mzcache"
	if result != expected {
		t.Errorf(stdTestMessage, result, expected)
	}
}
func TestReadMiss(t *testing.T) {
	t.Parallel()
	_, err := Read("cache_miss", 1)
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf(stdTestMessage, err, ErrCacheMiss.Error())
	}
}
func TestReadFilePath(t *testing.T) {
	t.Parallel()
	path, file, hashKey := getCacheFilePath("testpath")
	if path != getCacheDir()+"/fd/4f/" {
		t.Errorf("Cache path is wrong, %v", path)
	}
	if file != getCacheDir()+"/fd/4f/62f64cf7e327dc5a460e1c3ab20b097365438a74977da31d3e93b2299247.gz" {
		t.Errorf("Cache file is wrong, %v", file)
	}
	if hashKey != "fd4f62f64cf7e327dc5a460e1c3ab20b097365438a74977da31d3e93b2299247" {
		t.Errorf("hashKey is wrong, %v", hashKey)
	}
}

// run write/read cycles in parallel to validate locking works properly and there
// are no race conditions
func TestWriteUnderLoad(t *testing.T) {
	t.Parallel()
	errors := []error{}
	var wg sync.WaitGroup
	var errMutex sync.Mutex
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := testWrite()
			if err != nil {
				errMutex.Lock()
				errors = append(errors, err)
				errMutex.Unlock()
			}
		}()
	}
	wg.Wait()
	if len(errors) != 0 {
		t.Errorf("got unexpected errors, %v", errors)
	}
}
