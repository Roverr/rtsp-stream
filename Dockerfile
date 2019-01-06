## Build stage
FROM golang:1.11-alpine AS build-env
ADD ./main.go /go/src/github.com/Roverr/rtsp-stream/main.go
ADD ./core /go/src/github.com/Roverr/rtsp-stream/core
ADD ./Gopkg.lock /go/src/github.com/Roverr/rtsp-stream/Gopkg.lock
ADD ./Gopkg.toml /go/src/github.com/Roverr/rtsp-stream/Gopkg.toml
WORKDIR /go/src/github.com/Roverr/rtsp-stream
RUN apk add --update --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o server

## Creating potential production image
FROM alpine
RUN apk update && apk add bash ca-certificates ffmpeg && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build-env /go/src/github.com/Roverr/rtsp-stream/server /app/
ENTRYPOINT [ "/app/server" ]
