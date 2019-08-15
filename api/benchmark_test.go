package api

import (
	"math/rand"
	"strconv"
	"testing"
)

var maxFileSize = DefaultMaxFileSizeBytes

func BenchmarkLocalFileTransferEmptyFile(b *testing.B) {
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
		b.Fatal(err)
	}
	defer sync.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		createFileOrDie(alphaDir + "/" + strconv.Itoa(i))
		assertFileExists(b, betaDir+"/"+strconv.Itoa(i))
	}
}

func BenchmarkLocalFileTransferTenKBFile(b *testing.B) {
	benchmarkFileSize(b, 10*1000)
}

func BenchmarkLocalFileTransferOneHundredKB(b *testing.B) {
	benchmarkFileSize(b, 100*1000)
}

func BenchmarkLocalFileTransferOneMegabyte(b *testing.B) {
	benchmarkFileSize(b, 1000*1000)
}

func BenchmarkLocalFileTransferTenMegabyte(b *testing.B) {
	benchmarkFileSize(b, 10*1000*1000)
}

func BenchmarkLocalFileTransferOneHundredMegabyte(b *testing.B) {
	benchmarkFileSize(b, 100*1000*1000)
}

func BenchmarkMaxFileSizeOneMegabyte(b *testing.B) {
	maxFileSize = 1 * 1000
	defer func() { maxFileSize = DefaultMaxFileSizeBytes }()

	benchmarkFileSize(b, 100*1000*1000)
}

func benchmarkFileSize(b *testing.B, size int) {
	before()

	createDirOrDie(alphaDir)
	createDirOrDie(betaDir)

	sync, err := Sync(SyncOptions{
		Integration: "./sfs-localsync",
		Targets:     []string{alphaDir},
		Arguments: map[string]string{
			"destinations": betaDir,
		},
		MaxFileSizeBytes: maxFileSize,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer sync.Close()

	var contents []string
	for i := 0; i < b.N; i++ {
		random := make([]byte, size)
		rand.Read(random)
		contents = append(contents, string(random))
	}

	b.ResetTimer()

	for i, v := range contents {
		writeFileOrDie(alphaDir+"/"+strconv.Itoa(i), v)
		assertFileExists(b, betaDir+"/"+strconv.Itoa(i))
	}

}
