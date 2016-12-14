test:
	go test -v

cover:
	rm -rf *.coverprofile
	go test -coverprofile=request.coverprofile
	gover
	go tool cover -html=request.coverprofile