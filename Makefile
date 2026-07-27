.PHONY: build run dev clean daemon docker-up docker-down docker-logs hx-start hx-stop hx-status help

APP_NAME := football-wizard
CMD_DIR := ./cmd/football-wizard
OUTPUT := ./bin/$(APP_NAME)
SOURCES := $(shell find . -name "*.go" -not -path "./.agents/*")

build: $(OUTPUT)

$(OUTPUT): $(SOURCES)
	go build -o $(OUTPUT) $(CMD_DIR)

run: $(OUTPUT)
	$(OUTPUT)

dev:
	go run $(CMD_DIR)

daemon:
	go run $(CMD_DIR) daemon

clean:
	rm -rf bin/ data/ tmp/
	go clean -cache

docker-up:
	docker compose --env-file .env up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

hx-start:
	headlessx start

hx-stop:
	headlessx stop

hx-status:
	headlessx status

help:
	@echo "Usage:"
	@echo "  make build          Build binary"
	@echo "  make run            Run TUI (requires HeadlessX running)"
	@echo "  make dev            Go run directly"
	@echo "  make daemon         Run scheduler headless"
	@echo "  make clean          Clean artifacts"
	@echo ""
	@echo "  make docker-up      Start full stack (HeadlessX + app)"
	@echo "  make docker-down    Stop all containers"
	@echo "  make docker-logs    Tail logs"
	@echo ""
	@echo "  make hx-start       Start HeadlessX natively"
	@echo "  make hx-stop        Stop HeadlessX"
	@echo "  make hx-status      HeadlessX status"
