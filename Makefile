test: compile-binaries
	
	go test ./api/

compile-binaries:
	echo "Hello world"
	go build sfs-localsync/main.go
	mv main api/sfs-localsync

watch: compile-binaries
	modd
