DOCKER_VERSION?=2

test: ## Runs tests
	go test ./...
run:  ## Builds & Runs the application
	go build . && ./rtsp-stream
docker-build:  ## Builds normal docker container
	docker build -t roverr/rtsp-stream:${DOCKER_VERSION} .
docker-debug: ## Builds the image and starts it in debug mode
	rm -rf ./log && mkdir log && \
	$(MAKE) docker-build && \
	docker run -d \
	-v `pwd`/log:/var/log \
	-e RTSP_STREAM_DEBUG=true \
	-e RTSP_STREAM_AUTH_JWT_ENABLED=true \
	-e RTSP_STREAM_AUTH_JWT_SECRET=your-256-bit-secret \
	-e RTSP_STREAM_BLACKLIST_COUNT=2 \
	-p 8080:8080 \
	roverr/rtsp-stream:${DOCKER_VERSION}
docker-build-mg:  ## Builds docker container with management UI
	docker build -t roverr/rtsp-stream:${DOCKER_VERSION}-management -f Dockerfile.management .
docker-debug-mg: ## Builds management image and starts it in debug mode
	rm -rf ./log && mkdir log && \
	$(MAKE) docker-build-mg && \
	docker run -d \
	-v `pwd`/log:/var/log \
	-e RTSP_STREAM_DEBUG=true \
	-e RTSP_STREAM_BLACKLIST_COUNT=2 \
	-p 3000:80 -p 8080:8080 \
	roverr/rtsp-stream:${DOCKER_VERSION}-management
docker-all: ## Runs tests then builds all versions of docker images
	$(MAKE) test && $(MAKE) docker-build && $(MAKE) docker-build-mg
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
