.PHONY: build install test fmt vet check clean

build:
	go build -o gh-inline .

install: build
	gh extension install .

test:
	go test -race ./...

fmt:
	gofmt -l .

vet:
	go vet ./...

check: fmt vet test

clean:
	rm -f gh-inline
