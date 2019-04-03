# <img src="./rtsp-stream.png"/>

[![Go Report Card](https://goreportcard.com/badge/github.com/Roverr/rtsp-stream)](https://goreportcard.com/report/github.com/Roverr/rtsp-stream)
 [![Maintainability](https://api.codeclimate.com/v1/badges/202152e83296250ab527/maintainability)](https://codeclimate.com/github/Roverr/rtsp-stream/maintainability)

rtsp-stream is an easy to use out of box solution that can be integrated into existing systems resolving the problem of not being able to play rtsp stream natively in browsers. 

## Table of contents
* [How does it work](#how-does-it-work)
* [Authentication](#authentication)
    * [No Authentication](#no-authentication)
    * [JWT](#jwt-authentication)
* [Easy API](#easy-api)
* [Configuration](#configuration)
    * [Transcoding](#transcoding-related-configuration)
    * [HTTP](#http-related-configuration)
    * [CORS](#cors-related-configuration)
* [Run with Docker](#run-with-docker)
* [UI](#ui)
* [Proven players](#proven-players)
* [Coming soon features](#coming-soon-features)


## How does it work
It converts `RTSP` streams into `HLS` based on traffic. The idea behind this is that the application should not transcode anything until someone is actually watching the stream. This can help with network bottlenecks in systems where there are a lot of cameras installed.

There's a running go routine in the background that checks if a stream is being active or not. If it's not the transcoding stops until the next request for that stream.

## Authentication

The application offers different ways for authentication. There are situations when you can get away with no authentication, just
trusting requests because they are from reliable sources or just because they know how to use the API. In other cases, production cases, you definitely
want to protect the service. This application was not written to handle users and logins, so authentication is as lightweight as possible.


### No Authentication

**By default there is no authentication** what so ever. This can be useful if you have private subnets
where there is no real way to reach the service from the internet. (So every request is kind of trusted.) Also works great
if you just wanna try it out, maybe for home use.


### JWT Authentication

You can use shared key JWT authentication for the service.

The service itself does not create any tokens, but your authentication service can create.
After it's created it can be validated in the transcoder using the same secret / keys.
It is the easiest way to integrate into existing systems.

The following environment variables are available for this setup:

| Env variable | Description | Default | Type |
| :---        |    :----   |          ---: | :--- |
| RTSP_STREAM_AUTH_JWT_ENABLED | Indicates if the service should use the JWT authentication for the requests | `false` | bool |
| RTPS_STREAM_AUTH_JWT_SECRET | The secret used for creating the JWT tokens | `macilaci` | string |
| RTSP_STREAM_AUTH_JWT_PUB_PATH | Path to the public shared RSA key.| `/key.pub` | string |
| RTSP_STREAM_AUTH_JWT_METHOD | Can be `secret` or `rsa`. Changes how the application does the JWT verification.| `secret` | string |

You won't need the private key for it because no signing happens in this application.

<img src="./transcoder_auth.png"/>

## Easy API
**There are 2 main endpoints to call:**

`POST /start`

Starts the transcoding of the given stream. You have to pass URI format with rtsp procotol. 
The respond should be considered the subpath for the video player to call.
So if your applicaiton is `myapp.com` then you should call `myapp.com/stream/host/index.m3u8` in your video player.
The reason for this is to remain flexible regarding useability. 

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

`GET /list`

This (kind of a debug) endpoint is used to list the streams in the system. 
Since the application does not handle users, it does not handle permissions obviously. 
You might not want everyone to be able to list the streams 
available in the system. But if you do, you can use this. You just have to enable it via [env variable](https://github.com/Roverr/rtsp-stream#configuration).



Response:
```js
[
    {
        "running": true,
        "uri": "/stream/185.180.88.98-streaming-channels-101/index.m3u8"
    }
]
``` 

## Configuration

You can configure the following settings in the application with environment variables:

### Transcoding related configuration:

| Env variable | Description | Default | Type |
| :---        |    :----   |          ---: | :--- |
| RTSP_STREAM_CLEANUP_TIME | Time period for the cleanup process [info on format here](https://golang.org/pkg/time/#ParseDuration) | `2m0s` | string |
| RTSP_STREAM_STORE_DIR | Sub directory to store the video chunks | `./videos` | string |
| RTSP_STREAM_KEEP_FILES | Option to keep the chunks for the stream being transcoded | `false` | bool |

The project uses [Lumberjack](https://github.com/natefinch/lumberjack) for the log rotation of the ffmpeg transcoding processes.

| Env variable | Description | Default | Type |
| :---        |    :----   |          ---: | :--- |
| RTSP_STREAM_PROCESS_LOGGING_ENABLED | Indicates if logging of transcoding ffmpeg processes is enabled or not | `false` | bool |
| RTSP_STREAM_PROCESS_LOGGING_DIR | Describes the directory where the transcoding logs are stored | `/var/log/rtsp-stream` | string |
| RTSP_STREAM_PROCESS_LOGGING_MAX_SIZE | Maximum size of each log file in **megabytes** | `500` | integer |
| RTSP_STREAM_PROCESS_LOGGING_MAX_AGE | Maximum number of days that we store a given log file. | `7` | integer |
| RTSP_STREAM_PROCESS_LOGGING_MAX_BACKUPS | Maximum number of old log files to retain | `3` | integer |
| RTSP_STREAM_PROCESS_LOGGING_COMPRESS | Option to compress the rotated log file or not | `true` | bool |

<hr>

### HTTP related configuration:

| Env variable | Description | Default | Type |
| :---        |    :----   |          ---: | :--- |
| RTSP_STREAM_PORT | Port where the application listens | `8080` | integer |
| RTSP_STREAM_DEBUG | Turns on / off debug logging | `false` | bool |
| RTSP_STREAM_LIST_ENDPOINT | Turns on / off the `/list` endpoint | `false` | bool |

<hr>

### CORS related configuration

By default all origin is allowed to make requests to the server, but you might want to configure it for security reasons.

| Env variable | Description | Default | Type |
| :---        |    :----   |          ---: | :--- |
| RTSP_STREAM_CORS_ENABLED | Indicates if cors should be handled as configured or as default (everything allowed) | `false` | bool |
| RTSP_STREAM_CORS_ALLOWED_ORIGIN | A list of origins a cross-domain request can be executed from |  | []string |
| RTSP_STREAM_CORS_ALLOW_CREDENTIALS | Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates | `false` | bool |
| RTSP_STREAM_CORS_MAX_AGE | Indicates how long (in seconds) the results of a preflight request can be cached. | `0` | integer |

## Run with Docker
The application has an offical docker repository at dockerhub, therefore you can easily run it with simple commands:

```bash
docker run -p 80:8080 roverr/rtsp-stream:1
```

or you can build it yourself using the source code.

## UI

You can use the included UI for handling the streams. The UI is not a compact solution right now, but it gets the job done.

Running it with docker:

```sh
docker run -p 80:80 -p 8080:8080 roverr/rtsp-stream:1-management
```

If you decide to use the management image, you should know that port 80 is flexible, you can set it to whatever you prefer, but 8080 is currently burnt into the UI as the ultimate port of the backend.

You should expect something like this:


<img src="./ui.gif"/>


### Logs

With the UI solution the following files are created in `/var/log`: 
* rtsp-stream-ui.err.log
* rtsp-stream-ui.out.log
* rtsp-stream.err.log
* rtsp-stream.out.log

## Proven players
The following list of players has been already tried out in production environment using this backend:

* Angular - [videogular](http://www.videogular.com/)
* React - [ReactHLS](https://github.com/foxford/react-hls)

## Coming soon features

✅ - Done

🤷‍♂️ - Needs more labour

* ✅ Proper logging - File logging for the output of ffmpeg with the option of rotating file log
* ✅ Improved cleanup - Unused streams should be removed from the system after a while
* 🤷‍♂️ API improvements - Delete endpoint for streams so clients can remove streams whenever they would like to
* ✅  Authentication layer - More options for creating authentication within the service
