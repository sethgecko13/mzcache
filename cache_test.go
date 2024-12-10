package mzcache

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func testWrite() error {
	expected := "value1\nvalue2\nvalue3"
	key := "test1"
	err := Write(key, expected)
	if err != nil {
		return err
	}
	result, _ := Read(key, 1)
	if result != expected {
		return fmt.Errorf("result was incorrect, got: %v, want: %v.", result, expected)
	}
	return nil
}

func TestWriteCache(t *testing.T) {
	t.Parallel()
	err := testWrite()
	if err != nil {
		t.Errorf("got an unexpected error: %v", err.Error())
	}
}
func TestWriteCacheEmptyString(t *testing.T) {
	t.Parallel()
	value := ""
	key := "empty"
	err := Write(key, value)
	if err != ErrCacheEmptyString {
		t.Errorf("got %v, expected: %v", err.Error(), ErrCacheEmptyString.Error())
	}
}
func TestGetCacheExpired(t *testing.T) {
	t.Parallel()
	expected := "value1\nvalue2\nvalue3"
	key := "expired"
	err := Write(key, expected)
	if err != nil {
		t.Errorf("got unexpected error: %v", err.Error())
	}
	_, err = Read(key, -1)
	var expiredErr *ErrCacheExpired
	if !errors.As(err, &expiredErr) {
		t.Errorf("got unexpected error: %v", err.Error())
	}
}
func TestGetCacheHit(t *testing.T) {
	t.Parallel()
	expected := "value1\nvalue2\nvalue3"
	key := "test2"
	err := Write(key, expected)
	if err != nil {
		t.Errorf("got unexpected error: %v", err.Error())
	}
	result, err := Read(key, 1)
	if err != nil {
		t.Errorf("got unexpected error: %v", err.Error())
	}
	if result != expected {
		t.Errorf("result was incorrect, got: %v, want: %v.", result, expected)
	}
}
func TestGetCacheInvalidDirectory(t *testing.T) {
	// Don't run in parallel, will break other tests
	oldCachePath := cachePath
	cachePath = "/var/tmp/blah"
	defer func() {
		cachePath = oldCachePath
	}()
	_, err := Read("invalid_directory", 1)
	if !errors.Is(err, ErrCacheDirectoryNotExist) {
		t.Errorf("result was incorrect, got: %v, want: %v.", err, ErrCacheDirectoryNotExist.Error())
	}
}
func TestGetCacheMiss(t *testing.T) {
	t.Parallel()
	_, err := Read("cache_miss", 1)
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("Cache existed but shouldn't have")
	}
}
func TestGetCacheFilePath(t *testing.T) {
	t.Parallel()
	path, file, hashKey := getCacheFilePath("testpath")
	if path != cachePath+"/fd/4f/" {
		t.Errorf("Cache path is wrong, %v", path)
	}
	if file != cachePath+"/fd/4f/62f64cf7e327dc5a460e1c3ab20b097365438a74977da31d3e93b2299247.gz" {
		t.Errorf("Cache file is wrong, %v", file)
	}
	if hashKey != "fd4f62f64cf7e327dc5a460e1c3ab20b097365438a74977da31d3e93b2299247" {
		t.Errorf("hashKey is wrong, %v", hashKey)
	}
}
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
