.PHONY: build run clean tidy vet test \
        docker-build docker-up docker-down docker-down-v docker-restart docker-logs

build:
	go build -o bin/server ./cmd/server/

run: build
	./bin/server

clean:
	rm -rf bin/

tidy:
	go mod tidy

vet:
	go vet ./...

test:
	go test ./...

# ---- docker ----
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-down-v:
	docker compose down -v

docker-restart:
	docker compose down && docker compose up -d

docker-logs:
	docker compose logs -f
