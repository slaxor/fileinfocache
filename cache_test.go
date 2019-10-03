package fileinfocache

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testFileInfo struct{}

func (tfi testFileInfo) Name() string      { return "A Testfile that doesn't exist" }
func (tfi testFileInfo) Size() int64       { return 1234 }
func (tfi testFileInfo) Mode() os.FileMode { return 0777 }
func (tfi testFileInfo) ModTime() time.Time {
	t, _ := time.Parse("2006-01-02 15:06:00", "2019-10-03 01:21:54")
	return t
}
func (tfi testFileInfo) IsDir() bool      { return false }
func (tfi testFileInfo) Sys() interface{} { return nil }

func makeTestFile(fname string, c []byte) func() {
	if _, err := os.Stat(fname); err == nil {
		log.Fatalf("Your desired filename \"%s\" exists. Please choose another, or remove the file", fname)
	}
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.Write(c)
	return func() {
		err := os.Remove(fname)
		if err != nil {
			log.Fatalf("%s\nplease clean the file \"%s\" manually", err, fname)
		}
	}
}

func Test_NewFileInfo(t *testing.T) {
	tfi := testFileInfo{}
	result := NewFileInfo(tfi)
	t.Run("has a Name", func(t *testing.T) {
		expect := tfi.Name()
		assert.Equal(t, expect, result.Name)
	})
	t.Run("has a Size", func(t *testing.T) {
		expect := tfi.Size()
		assert.Equal(t, expect, result.Size)
	})
	t.Run("has a Mode", func(t *testing.T) {
		expect := tfi.Mode()
		assert.Equal(t, expect, result.Mode)
	})
	t.Run("has a ModTime", func(t *testing.T) {
		expect := tfi.ModTime()
		assert.Equal(t, expect, result.ModTime)
	})
	t.Run("has a IsDir", func(t *testing.T) {
		expect := tfi.IsDir()
		assert.Equal(t, expect, result.IsDir)
	})
}

// func Test_CacheFromDir(t *testing.T) {
// }

func Test_CacheFromDirRecursive(t *testing.T) {
	result := CacheFromDirRecursive(".")
	assert.IsType(t, Cache{}, result)
}

func Test_NewCache(t *testing.T) {
	fname := "Foo"
	rmTestFile := makeTestFile(fname, []byte("Some Content"))
	defer rmTestFile()
	tfis := FileInfos{FileInfo{Name: fname}}
	result := NewCache(tfis)
	assert.IsType(t, Cache{}, result)
}

func Test_Cache_makeKey(t *testing.T) {
	fname := "Foo"
	rmTestFile := makeTestFile(fname, []byte("Some Content"))
	defer rmTestFile()
	tfi := FileInfo{Name: fname}
	c := Cache{}
	result := c.makeKey(tfi)
	assert.Equal(t, "78138d2003f1a87043d65c692fb3a64b", result)

}

func Test_Cache_Insert(t *testing.T) {
	fname := "Foo"
	rmTestFile := makeTestFile(fname, []byte("Some Content"))
	defer rmTestFile()
	tfi := FileInfo{Name: fname}
	c := Cache{}
	c.Insert(tfi)
	result := c["78138d2003f1a87043d65c692fb3a64b"][0].Name
	assert.Equal(t, fname, result)

}

func Test_Cache_Write(t *testing.T) {
	fname := "Foo"
	rmTestFile := makeTestFile(fname, []byte("Some Content"))
	defer rmTestFile()
	tfi := FileInfo{Name: fname}
	c := Cache{}
	c.Insert(tfi)
	cname := "testCacheFile.json.gz"
	c.Write(cname)
	defer os.Remove(cname)
	cmd := exec.Command("gzip", "-dc")
	cio, err := os.Open(cname)
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdin = cio
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	expect := `{"78138d2003f1a87043d65c692fb3a64b":[{"name":"Foo","size":0,"mode":0,"modTime":"0001-01-01T00:00:00Z","isDir":false}]}`
	assert.Equal(t, expect, out.String())
}

func Test_CacheFromFile(t *testing.T) {
	cacheFileContent := []byte{
		0x1f, 0x8b, 0x08, 0x18, 0x4e, 0x5a, 0x95, 0x5d,
		0x00, 0xff, 0x63, 0x61, 0x63, 0x68, 0x65, 0x2e,
		0x6a, 0x73, 0x6f, 0x6e, 0x00, 0x41, 0x20, 0x63,
		0x61, 0x63, 0x68, 0x65, 0x20, 0x66, 0x69, 0x6c,
		0x65, 0x20, 0x66, 0x6f, 0x72, 0x20, 0x67, 0x69,
		0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
		0x2f, 0x73, 0x6c, 0x61, 0x78, 0x6f, 0x72, 0x2f,
		0x64, 0x65, 0x64, 0x75, 0x70, 0x69, 0x66, 0x69,
		0x65, 0x72, 0x00, 0x3c, 0xca, 0xb1, 0x0a, 0xc2,
		0x30, 0x14, 0x85, 0xe1, 0x77, 0x39, 0x73, 0x84,
		0x93, 0xa6, 0xa6, 0xf1, 0xce, 0xe2, 0x13, 0x74,
		0x52, 0x1c, 0x52, 0x93, 0xc0, 0x05, 0x6b, 0xc0,
		0x6c, 0x96, 0xbe, 0xbb, 0x90, 0xa1, 0xf0, 0x0d,
		0xff, 0xf0, 0x6f, 0x98, 0x82, 0x75, 0x21, 0x0d,
		0xa4, 0x2b, 0x36, 0x86, 0x89, 0xa3, 0x4b, 0xfe,
		0xfc, 0xf2, 0x97, 0xa1, 0x2c, 0x2e, 0xfa, 0x71,
		0x81, 0x3c, 0x36, 0x7c, 0xe2, 0x9a, 0x21, 0xb8,
		0xd5, 0x0a, 0x83, 0xa6, 0xbf, 0x0c, 0xa1, 0xc1,
		0x5a, 0xd3, 0x11, 0xb3, 0xf6, 0x83, 0xa4, 0x3d,
		0x75, 0x33, 0x29, 0xdd, 0x1d, 0x06, 0xda, 0xae,
		0xfa, 0x85, 0x94, 0xf8, 0x6e, 0x79, 0x7f, 0xee,
		0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0xe2, 0xc3,
		0x3f, 0x9d, 0x76, 0x00, 0x00, 0x00,
	}
	fname := "testCacheFile.json.gz"
	rmTestFile := makeTestFile(fname, cacheFileContent)
	defer rmTestFile()
	c := CacheFromFile(fname)
	expect := "Foo"
	assert.Equal(t, expect, c["78138d2003f1a87043d65c692fb3a64b"][0].Name)
}
