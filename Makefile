BINARY   := bin/football-wizard
MODULE   := github.com/edorguez/football-wizard
GOFLAGS  := -ldflags="-s -w"
GOTAGS   :=

.PHONY: build run dev test clean

# Build the binary into bin/football-wizard
#   make build
build:
	@mkdir -p bin data
	go build $(GOFLAGS) -o $(BINARY) ./cmd/football-wizard

# Build and run the binary
#   make run
#   make run ARGS="--season 2025"
run: build
	./$(BINARY) $(ARGS)

# Build and run via go run (faster for development)
#   make dev                          # launch the TUI
#   make dev ARGS="daemon"            # headless scheduler
#   make dev ARGS="--season 2025"     # scrape a season
#   make dev ARGS="--season 2025 --full"
#   make dev ARGS="--season 2025 --full --workers 5"
#   make dev ARGS="--train"
dev:
	@mkdir -p data
	go run ./cmd/football-wizard $(ARGS)

# Run all tests with race detection
#   make test
test:
	go test -v -race -count=1 ./...

# Remove bin/ artifacts and data/ (SQLite DB + HTML cache)
#   make clean
clean:
	rm -rf bin/
	rm -rf data/

# Catch-all: prevent make errors from stray args
%:
	@:
