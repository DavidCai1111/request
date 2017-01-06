test:
	go test -v -race

cover:
	rm -rf *.coverprofile
	go test -coverprofile=request.coverprofile -v -race
	gover
	go tool cover -html=request.coverprofile
	rm -rf *.coverprofile
