test: ## Runs tests
	go test ./...
run:  ## Builds & Runs the application
	go build . && ./rtsp-stream
docker-build:  ## Builds normal docker container
	docker build -t roverr/rtsp-stream:1 .
docker-build-mg:  ## Builds docker container with management UI
	docker build -t roverr/rtsp-stream:1-management -f Dockerfile.management .

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
