APP_NAME := go-server
IMAGE_TAG ?= local

.PHONY: help fmt vet test build docker-build docker-run kind-up kind-load k8s-apply

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

fmt:
	@test -z "$$(gofmt -l .)" || (echo "Run 'gofmt -w .' to fix formatting:" && gofmt -l . && exit 1)

vet:
	go vet ./...

test:
	go test -v ./internal/...

build:
	go build -trimpath -ldflags="-w -s" -o bin/server ./cmd/server

docker-build:
	docker build -t $(APP_NAME):$(IMAGE_TAG) .

docker-run:
	docker compose up --build -d

docker-down:
	docker compose down
