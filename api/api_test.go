package api

import (
	"fmt"
	"io/ioutil"
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

func assertFileExists(t *testing.T, path string) {
	var err error
	for i := 0; i < 3; i++ {
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

func assertFileContents(t *testing.T, path string, contents string) {
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	if contents != string(bytes) {
		t.Fatal("file contents wrong")
	}
}

func TestItCanSyncOneDirectoryToAnother(t *testing.T) {
	before()
	fmt.Println("0")

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

/*
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

func TestFilesCanBeReplacedWithPlaceholders(t *testing.T) {
	t.Skip()
}

func TestPlaceholderConversionCanBeConfigured(t *testing.T) {
	t.Skip()
}

func TestPlaceholdersCanBeRedownloaded(t *testing.T) {
	t.Skip()
}
*/
