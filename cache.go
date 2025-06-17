package mzcache

import (
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

var LockPath = getLockPath()

var ErrCacheEmptyString = errors.New("empty string passed in to cache")
var ErrCacheCreateDirectory = errors.New("unable to create cache directory")
var ErrCacheCreate = errors.New("unable to create cache file")
var ErrCacheWrite = errors.New("unable to write to cache file")
var ErrCacheSync = errors.New("unable to sync cache file to filesystem")

var ErrCacheLock = errors.New("unable to lock cache file")
var ErrCacheUnlock = errors.New("unable to unlock cache file")

var ErrCacheMiss = errors.New("cache file does not exist")
var ErrCacheDecompress = errors.New("unable to decompress cache file")
var ErrCacheRead = errors.New("unable to read cache file")

// all the other errors we can use errors.Join to give details about the error.
// with expired cache, we have to handle separately because there is no underlying error.
type ErrCacheExpired struct {
	FullPath string
}

func (e *ErrCacheExpired) Error() string {
	return fmt.Sprintf("cache expired %s", e.FullPath)
}

// since caching is so fundamental to my app, I choose to panic if caching does not work.
// you may want to make different decisions if you use this library.
func getLockPath() string {
	err := os.RemoveAll(getCacheDir() + "mz*")
	if err != nil {
		panic("unable to remove previous lock files")
	}
	dname, err := os.MkdirTemp(getCacheDir(), "mz")
	if err != nil {
		panic("unable to create cache lock file")
	}
	return dname
}
func hash(key string) string {
	h := sha256.New()
	h.Write([]byte(key))
	bs := h.Sum(nil)
	result := fmt.Sprintf("%x", bs)
	return result
}
func getCacheFilePath(key string) (path string, fullPath string, hashKey string) {
	hashKey = fmt.Sprintf("%v", hash(key))
	path = fmt.Sprintf("%v/%v/%v/", getCacheDir(), hashKey[0:2], hashKey[2:4])
	fullPath = fmt.Sprintf("%v%v.gz", path, hashKey[4:64])
	return path, fullPath, hashKey
}
func getFileLockPath(hashKey string) string {
	return fmt.Sprintf("%s/%s", LockPath, hashKey)
}
func Write(key string, value string) error {
	// don't write empty values they are probably errors
	if value == "" {
		return ErrCacheEmptyString
	}
	path, fullPath, hashKey := getCacheFilePath(key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0750)
		if err != nil {
			return errors.Join(ErrCacheCreateDirectory, err)
		}
	}
	fileLock := flock.New(getFileLockPath(hashKey))
	err := fileLock.Lock()
	if err != nil {
		return errors.Join(ErrCacheLock, err)
	}
	var lockError error
	defer func() {
		err = fileLock.Unlock()
		if err != nil {
			lockError = errors.Join(ErrCacheUnlock, err)
		}
	}()
	// no need to remove file, this causes race conditions, create removes previous contents
	file, err := os.Create(filepath.Clean(fullPath))
	if err != nil {
		return errors.Join(ErrCacheCreate, err)
	}
	defer file.Close()
	g := gzip.NewWriter(file)
	defer g.Close()
	_, err = g.Write([]byte(value))
	if err != nil {
		return errors.Join(ErrCacheWrite, err)
	}
	err = file.Sync()
	if err != nil {
		return errors.Join(ErrCacheSync, err)
	}
	return lockError
}
func Read(key string, days int) (string, error) {
	var result string
	path, fullPath, hashKey := getCacheFilePath(key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return result, errors.Join(ErrCacheMiss, err)
	}
	fileLock := flock.New(getFileLockPath(hashKey))
	err := fileLock.Lock()
	if err != nil {
		return result, errors.Join(ErrCacheLock, err)
	}
	var lockError error
	defer func() {
		err = fileLock.Unlock()
		if err != nil {
			lockError = errors.Join(ErrCacheUnlock, err)
		}
	}()
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return result, errors.Join(ErrCacheMiss, err)
	} else {
		file, _ := os.Stat(fullPath)
		now := time.Now()
		dt := now.AddDate(0, 0, ((days - 1) * -1))
		cutoff := dt.Format("2006-01-02")
		fileDate := file.ModTime()
		fd := fileDate.Format("2006-01-02")
		if fd < cutoff {
			return result, &ErrCacheExpired{FullPath: fullPath}
		}
		inputFile, err := os.Open(filepath.Clean(fullPath))
		if err != nil {
			return result, errors.Join(ErrCacheRead, err)
		}
		defer inputFile.Close()
		reader, err := gzip.NewReader(inputFile)
		if err != nil {
			return result, errors.Join(ErrCacheDecompress, err)
		}
		defer reader.Close()
		content, err := io.ReadAll(reader)
		if err != nil {
			return result, errors.Join(ErrCacheRead, err)
		}
		result = string(content)
		return result, lockError
	}
}
func getCacheDir() string {
	if cacheDir := os.Getenv("MZ_CACHE_DIR"); cacheDir != "" {
		return cacheDir
	}
	return "/var/tmp/mzcache"
}
