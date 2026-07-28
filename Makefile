BINARY   := bin/football-wizard
MODULE   := github.com/edorguez/football-wizard
GOFLAGS  := -ldflags="-s -w"
GOTAGS   :=

.PHONY: build run dev test clean

build:
	@mkdir -p bin data
	go build $(GOFLAGS) -o $(BINARY) ./cmd/football-wizard

run: build
	./$(BINARY) $(ARGS)

dev:
	@mkdir -p data
	go run ./cmd/football-wizard $(ARGS)

test:
	go test -v -race -count=1 ./...

clean:
	rm -rf bin/
	rm -rf data/

# Catch-all: prevent make errors from stray args
%:
	@:  # ignore unknown targets
