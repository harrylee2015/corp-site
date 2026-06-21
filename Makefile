.PHONY: build run clean tidy vet test css \
        docker-build docker-up docker-down docker-down-v docker-restart docker-logs

build:
	go build -o bin/server ./cmd/server/

# Regenerate the precompiled Tailwind CSS (requires Node/npx with network access).
# Output web/static/css/tailwind.css is committed so production needs no CDN.
css:
	npx tailwindcss@3.4.17 -c ./tailwind.config.js -i ./web/static/css/tailwind.input.css -o ./web/static/css/tailwind.css --minify

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
