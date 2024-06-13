lint:
	go fmt ./...
	go vet ./...
	golangci-lint run -v ./...

cloudci-build: lint
	GOARCH=darwin GOARCH=arm64 go build -o cloudci

cloudci-docker:
	docker build -t cloudci .

cloudci: build
	./cloudci