# <img src="./rtsp-stream.png"/>

rtsp-stream is an easy to use out of box solution that can be integrated into existing systems resolving the problem of not being able to play rtsp stream natively in browsers. 

## How does it work
It converts `RTSP` streams into `HLS` based on traffic. The idea behind this is that the application should not transcode anything until someone is actually watching the stream. This can help with network bottlenecks in systems where there are a lot of cameras installed.

There's a running go routine in the background that checks if a stream is being active or not. If it's not the transcoding stops until the next request for that stream.

## Easy API
There are 2 endpoints to call:

`POST /start`

Requires payload: `{ "uri": "rtsp://username:password@host" }`

Response: `{ "uri": "/stream/host/index.m3u8" }`


`GET /stream/host/*file`

Simple static file serving which is used when fetching chunks of `HLS`

## Configuration

You can configure the following settings in the application with environment variables:

* `RTSP_STREAM_CLEANUP_TIME` - Time period for the cleanup process in `ms`
* `RTSP_STREAM_STORE_DIR` - Sub directory to store video chunks
* `RTPS_STREAM_PORT` - Port where the application listens

## Run with Docker
The application has an offical docker repository at dockerhub, therefore you can easily run it with simple commands:

`docker run -p 80:8080 roverr/rtsp-stream`

or you can build it yourself using the source code.

