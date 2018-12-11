test:
	go test ./...
run:
	go build . && ./rtsp-stream
docker-build:
	docker build -t "roverr/rtsp-stream" .