## Build stage
FROM golang:1.11-alpine AS build-env
ADD . /go/src/github.com/Roverr/rtsp-stream
WORKDIR /go/src/github.com/Roverr/rtsp-stream
RUN apk add --update --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o server

## Creating potential production image
FROM alpine
RUN apk update && apk add ca-certificates ffmpeg && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build-env /go/src/github.com/Roverr/rtsp-stream/server /app/
ENTRYPOINT [ "/app/server" ]