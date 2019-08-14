clean:
	rm -rf ./api/.test_data

test: compile-binaries
	go test ./api/

test-only: compile-binaries
	go test -run=${NAME} ./api/

benchmark: compile-binaries
	go test ./api -run=XXX -bench=.

compile-binaries: clean
	go build sfs-localsync/main.go
	mv main api/sfs-localsync

watch: compile-binaries
	modd
