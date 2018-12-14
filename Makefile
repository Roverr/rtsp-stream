test:
	go test ./...
run:
	go build . && ./rtsp-stream
docker-build:
	docker build -t roverr/rtsp-stream:1 .
docker-build-mg:
	docker build -t roverr/rtsp-stream:1-management -f Dockerfile.management .
