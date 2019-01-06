test: ## Runs tests
	go test ./...
coverage: ## Runs tests with coverage going into cover.out
	go test ./... -coverprofile cover.out
open-coverage:
	go tool cover -html=cover.out
run:  ## Builds & Runs the application
	go build . && ./rtsp-stream
docker-build:  ## Builds normal docker container
	docker build -t roverr/rtsp-stream:1 .
docker-build-mg:  ## Builds docker container with management UI
	docker build -t roverr/rtsp-stream:1-management -f Dockerfile.management .
docker-all: ## Runs tests then builds all versions of docker images
	$(MAKE) test && $(MAKE) docker-build && $(MAKE) docker-build-mg
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
