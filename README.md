# <img src="./rtsp-stream.png"/>

rtsp-stream is an easy to use out of box solution that can be integrated into existing systems resolving the problem of not being able to play rtsp stream natively in browsers. 

## How does it work
It converts `RTSP` streams into `HLS` based on traffic. The idea behind this is that the application should not transcode anything until someone is actually watching the stream. This can help with network bottlenecks in systems where there are a lot of cameras installed.

There's a running go routine in the background that checks if a stream is being active or not. If it's not the transcoding stops until the next request for that stream.

## Easy API
**There are 2 main endpoints to call:**

`POST /start`

Requires payload:
```js
{ "uri": "rtsp://username:password@host" }
```

Response:
```js
{ "uri": "/stream/host/index.m3u8" }
```
<hr>

`GET /stream/host/*file`

Simple static file serving which is used when fetching chunks of `HLS`. This will be called by the client (browser) to fetch the chunks of the stream based on the given `index.m3u8`
<hr>
And there is also a third one which can be used for debugging (but you have to enable it via env variable)

`GET /list`

Lists all streams that are stored in the system along with their state of running:
```js
[
    {
        "running": true,
        "path": "185.180.88.98-streaming-channels-101"
    }
]
``` 
<hr>

`DELETE /stream/host`

Deletes a given host if it's a known stream.
Requires no payload.

<hr>


## Configuration

You can configure the following settings in the application with environment variables:

* `RTSP_STREAM_CLEANUP_TIME` - bool - Time period for the cleanup process [info on format here](https://golang.org/pkg/time/#ParseDuration) default: `2m0s`
* `RTSP_STREAM_STORE_DIR` - string - Sub directory to store video chunks
* `RTSP_STREAM_PORT` - number - Port where the application listens
* `RTSP_STREAM_DEBUG` - bool -Turns on / off debug logging
* `RTSP_STREAM_LIST_ENDPOINT` - bool -Turns on / off the `/list` endpoint (can be useful for debugging)

**CORS related configuration:**

By default all origin is allowed to make requests to the server, but you might want to configure it for security reasons.
* `RTSP_STREAM_CORS_ENABLED` - bool - Indicates if cors should be handled as configured or as default (everything allowed)
* `RTSP_STREAM_CORS_ALLOWED_ORIGIN` - string array - A list of origins a cross-domain request can be executed from
* `RTSP_STREAM_CORS_ALLOW_CREDENTIALS` - bool - Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates
* `RTSP_STREAM_CORS_MAX_AGE` - number - Indicates how long (in seconds) the results of a preflight request can be cached.

## Run with Docker
The application has an offical docker repository at dockerhub, therefore you can easily run it with simple commands:

`docker run -p 80:8080 roverr/rtsp-stream:1`

or you can build it yourself using the source code.

## UI

You can use the included UI for handling the streams. The UI is not a compact solution right now, but it gets the job done.

Running it with docker:

`docker run -p 80:80 -p 8080:8080 roverr/rtsp-stream:1-ui`
