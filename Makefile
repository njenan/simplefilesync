test: compile-binaries
	go test ./api/

test-only: compile-binaries
	go test -run=${NAME} ./api/

benchmark: compile-binaries
	go test ./api -run=XXX -bench=.

compile-binaries:
	go build sfs-localsync/main.go
	mv main api/sfs-localsync

watch: compile-binaries
	modd
