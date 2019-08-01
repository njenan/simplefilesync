package api

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var testDir string
var alphaDir string
var betaDir string
var deltaDir string

func init() {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testDir = basepath + "/.test_data"
	alphaDir = testDir + "/alpha"
	betaDir = testDir + "/beta"
	deltaDir = testDir + "/delta"
}

func before() {
	err := os.RemoveAll(testDir)
	if err != nil {
		panic(err)
	}

	err = os.Mkdir(testDir, 0777)
	if err != nil {
		panic(err)
	}
}

func createDirOrDie(path string) {
	err := os.Mkdir(path, 0777)
	if err != nil {
		panic(err)
	}
}

func createFileOrDie(path string) *os.File {
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	return file
}

func writeFileOrDie(path string, contents string) *os.File {
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	_, err = file.Write([]byte(contents))
	if err != nil {
		panic(err)
	}

	return file
}

func deleteFileOrDie(path string) {
	err := os.Remove(path)
	if err != nil {
		panic(err)
	}
}

type Fataler interface {
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

func assertFileExists(t Fataler, path string) {
	var err error
	for i := 0; i < 10; i++ {
		_, err = os.Stat(path)
		if err == nil {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	if err != nil {
		t.Fatal(err)
	}
}

func assertFileNotExists(t Fataler, path string) {
	var err error
	for i := 0; i < 10; i++ {
		_, err = os.Stat(path)
		if err != nil {
			fmt.Println(err)
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	if err == nil {
		t.Fatal("file existed")
	}
}

func assertFileContents(t Fataler, path string, contents string) {
	var err2 error

	for i := 0; i < 10; i++ {
		err := func() error {
			file, err := os.Open(path)
			if err != nil {
				return err
			}

			bytes, err := ioutil.ReadAll(file)
			if err != nil {
				return err
			}

			if contents != string(bytes) {
				return err
			}

			return nil
		}()
		if err != nil {
			err2 = err
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	if err2 != nil {
		t.Fatal(err2)
	}
}

func TestItCanSyncOneDirectoryToAnother(t *testing.T) {
	before()

	// create 2 directories
	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)

	// pipe from one to another
	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": betaDir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	createFileOrDie(alphaDir + "/one")
	assertFileExists(t, betaDir+"/one")

	createFileOrDie(alphaDir + "/two")
	assertFileExists(t, betaDir+"/two")
}

func TestItCanSyncFromOneDirectoryToMultiple(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)
	createDirOrDie(deltaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": betaDir + "," + deltaDir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	createFileOrDie(alphaDir + "/one")
	assertFileExists(t, betaDir+"/one")
	assertFileExists(t, deltaDir+"/one")
}

func TestFileContentsAreCopied(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": betaDir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	writeFileOrDie(alphaDir+"/one", "asdf")

	assertFileExists(t, betaDir+"/one")
	assertFileContents(t, betaDir+"/one", "asdf")
}
func TestMultipleTargetsCanBeSpecified(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)
	createDirOrDie(deltaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir, betaDir},
		Arguments: map[string]string{
			"destinations": deltaDir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	createFileOrDie(alphaDir + "/one")
	createFileOrDie(betaDir + "/two")

	assertFileExists(t, deltaDir+"/alpha/one")
	assertFileExists(t, deltaDir+"/beta/two")
}

func TestSubdirectoriesAreRecursivelyAdded(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(alphaDir + "/sub1")
	createDirOrDie(deltaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": deltaDir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	createFileOrDie(alphaDir + "/one")
	createFileOrDie(alphaDir + "/sub1/two")

	assertFileExists(t, deltaDir+"/one")
	assertFileExists(t, deltaDir+"/sub1/two")
}

func TestDeepSubdirectoriesAreRecursivelyAdded(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(alphaDir + "/sub1")
	createDirOrDie(alphaDir + "/sub1/sub2")
	createDirOrDie(deltaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": deltaDir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	createFileOrDie(alphaDir + "/one")
	createFileOrDie(alphaDir + "/sub1/sub2/two")

	assertFileExists(t, deltaDir+"/one")
	assertFileExists(t, deltaDir+"/sub1/sub2/two")
}

func TestItRemovesFilesWhenRemoveFilesIsSet(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": deltaDir,
		},
		RemoveDeletedFiles: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	createFileOrDie(alphaDir + "/one")

	assertFileExists(t, deltaDir+"/one")

	deleteFileOrDie(alphaDir + "/one")

	assertFileNotExists(t, deltaDir+"/one")
}

func TestVeryLargeFilesAreWrittenCorrectly(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": betaDir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	contents := make([]byte, 1000*1000+1)
	rand.Read(contents)

	writeFileOrDie(alphaDir+"/one", string(contents))

	assertFileExists(t, betaDir+"/one")
	assertFileContents(t, betaDir+"/one", string(contents))
}

func TestFilesCanBeReplacedWithPlaceholders(t *testing.T) {
	t.Skip()
}

func TestPlaceholderConversionCanBeConfigured(t *testing.T) {
	t.Skip()
}

func TestPlaceholdersCanBeRedownloaded(t *testing.T) {
	t.Skip()
}

func TestDirSubtraction(t *testing.T) {
	res, _ := subTargetsFromDir([]string{"/a/s/d/f", "/a/s/d/f"}, "/a/s/d/f")
	if res != "/" {
		t.Fatalf("res was %v not /", res)
	}

	res, _ = subTargetsFromDir([]string{"/a/s/d/f", "/a/s/d/e"}, "/a/s/d/f")
	if res != "/f" {
		t.Fatalf("res was %v not /f", res)
	}

	res, _ = subTargetsFromDir([]string{"/a/s/d/f"}, "/a/s/d/f/g")
	if res != "/g" {
		t.Fatalf("res was %v not /g", res)
	}

	res, _ = subTargetsFromDir([]string{"/a/s/d/f", "/a/s/d/e"}, "/a/s/d/f/g")
	if res != "/f/g" {
		t.Fatalf("res was %v not /f/g", res)
	}
}
