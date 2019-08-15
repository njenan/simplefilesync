package api

import (
	"testing"
)

func TestFilesCanBeReplacedWithPlaceholders(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": betaDir,
		},
		UsePlaceholders: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	writeFileOrDie(alphaDir+"/one", string("asdf"))

	assertFileExists(t, alphaDir+"/one.placeholder")

	assertFileExists(t, betaDir+"/one")
	assertFileContents(t, betaDir+"/one", string("asdf"))
}

func TestPlaceholdersCanBeDeletedAndDeleteTheirSourceFile(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": betaDir,
		},
		UsePlaceholders: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	writeFileOrDie(alphaDir+"/one", string("asdf"))

	assertFileExists(t, alphaDir+"/one.placeholder")
	assertFileExists(t, betaDir+"/one")

	deleteFileOrDie(alphaDir + "/one.placeholder")
	assertFileNotExists(t, betaDir+"/one")
}

func TestItHandlesFilesThatHavePlaceholderInTheirNameButNotTheirExtension(t *testing.T) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": betaDir,
		},
		UsePlaceholders: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Close()

	writeFileOrDie(alphaDir+"/one.placeholder.one", string("asdf"))

	assertFileExists(t, alphaDir+"/one.placeholder.one.placeholder")
	assertFileExists(t, betaDir+"/one.placeholder.one")

	deleteFileOrDie(alphaDir + "/one.placeholder.one.placeholder")
	assertFileNotExists(t, betaDir+"/one.placeholder.one")
}

func TestPlaceholdersCanBeRedownloaded(t *testing.T) {
	t.Skip()
}
