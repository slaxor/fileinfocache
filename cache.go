package fileinfocache

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// FileInfo is the static representation of an os.FileInfo
type FileInfo struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"modTime"`
	IsDir   bool        `json:"isDir"`
}

// NewFileInfo creates a FileInfo from an os.FileInfo
func NewFileInfo(fi os.FileInfo) FileInfo {
	return FileInfo{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
	}
}

// FileInfos is a collection of FileInfo
type FileInfos []FileInfo

// Cache is a collection of FileInfos accessable via a calculated key
type Cache map[string]FileInfos

// Not needed:
// CacheFromDir creates a Cache from a startDir
// func CacheFromDir(startDir string, recursive bool) Cache {
//     dir, err := os.Open(startDir)
//     if err != nil {
//         log.Fatal(err)
//     }
//     entries, err := dir.Readdir(0)
//     if err != nil {
//         log.Fatal(err)
//     }
//     fis := make(FileInfos, 1000)
//     return NewCache(fis)
// }

// CacheFromDirRecursive creates a Cache from a startDir and all subdirs
func CacheFromDirRecursive(startDir string) Cache {
	fis := make(FileInfos, 0, 1000)
	fullPath, err := filepath.Abs(startDir)
	if err != nil {
		log.Fatal(err)
	}
	filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		fi := NewFileInfo(info)
		fi.Name = path
		fis = append(fis, fi)
		clen := len(fis)
		if clen%1000 == 0 {
			log.Printf("scanned %d files", clen)
		}
		return err
	})

	cache := NewCache(fis)
	return cache
}

// NewCache takes a collection of os.FileInfo and inserts each in the
// newly created Cache
func NewCache(entries FileInfos) Cache {
	c := make(Cache, 1000)
	for _, e := range entries {
		c.Insert(e)
	}

	return c
}

// Insert takes a FileInfo and puts it in the appropriate position in
// the Cache
func (c Cache) Insert(fi FileInfo) {
	key := c.makeKey(fi)
	c[key] = append(c[key], fi)
}

// makeKey calculates the md5sum of the referenced file for a key
func (c Cache) makeKey(fi FileInfo) string {
	f, err := os.Open(fi.Name)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", (h.Sum(nil)))
}

// Write stores the Cache in a File
func (c Cache) Write(fname string) {
	output, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Name = "cache.json"
	zw.Comment = "A cache file for github.com/slaxor/fileinfocache"
	zw.ModTime = time.Now()
	_, err = zw.Write(output)
	if err != nil {
		log.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.Write(buf.Bytes())
}

// CacheFromFile reads a file and creates a Cache
func CacheFromFile(fname string) Cache {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	zr, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	fmt.Printf("Name: %s\nComment: %s\nModTime: %s\n\n", zr.Name, zr.Comment, zr.ModTime.UTC())

	if _, err := io.Copy(&buf, zr); err != nil {
		log.Fatal(err)
	}
	if err := zr.Close(); err != nil {
		log.Fatal(err)
	}

	var cache Cache
	err = json.Unmarshal(buf.Bytes(), &cache)
	if err != nil {
		log.Fatal(err)
	}
	return cache
}
