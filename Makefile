.PHONY: runtime

VERSION = "$(shell git rev-parse --short HEAD)"
DATE = "$(shell date +%m%d%H%M%S%N)"

build:
	mkdir -p bin
	go build -ldflags "-X main.version=0.9.X-$(VERSION)-$(DATE)-$(USER)" -o bin/nunu ./cmd/nunu

install: build
	sudo mv bin/nunu /usr/local/bin/Forum

test:
	go get -d ./cmd/dabber
	go test ./cmd/dabber

clean:
	rm -f nunu
