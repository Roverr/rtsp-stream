# <img src="./rtsp-stream.png"/>

[![Go Report Card](https://goreportcard.com/badge/github.com/Roverr/rtsp-stream)](https://goreportcard.com/report/github.com/Roverr/rtsp-stream)
 [![Maintainability](https://api.codeclimate.com/v1/badges/202152e83296250ab527/maintainability)](https://codeclimate.com/github/Roverr/rtsp-stream/maintainability)
 [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
 ![GitHub last commit](https://img.shields.io/github/last-commit/Roverr/rtsp-stream.svg)
 ![GitHub release](https://img.shields.io/github/release/Roverr/rtsp-stream.svg)

rtsp-stream is an easy to use out of box solution that can be integrated into existing systems resolving the problem of not being able to play rtsp stream natively in browsers. 

## Table of contents
* [How does it work](#how-does-it-work)
* [Run with Docker](#run-with-docker)
* [Authentication](#authentication)
    * [No Authentication](#no-authentication)
    * [JWT](#jwt-authentication)
* [Easy API](#easy-api)
* [Configuration](#configuration)
    * [Transcoding](#transcoding-related-configuration)
    * [HTTP](#http-related-configuration)
    * [CORS](#cors-related-configuration)
* [UI](#ui)
* [Debug](#debug)
* [Proven players](#proven-players)
* [Coming soon](#coming-soon)


## How does it work
It converts `RTSP` streams into `HLS` based on traffic. The idea behind this is that the application should not transcode anything until someone is actually watching the stream. This can help with network bottlenecks in systems where there are a lot of cameras installed.

There's a running go routine in the background that checks if a stream is being active or not. If it's not the transcoding stops until the next request for that stream.

## Run with Docker
The application has an offical [Docker repository](https://hub.docker.com/r/roverr/rtsp-stream/) at Dockerhub, therefore you can easily run it with simple commands:

```s
docker run -p 80:8080 roverr/rtsp-stream:1
```

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

API consisents of 2 main endpoints and one more extending them for debug purposes. [Read more](docs/api/README.md).

## Configuration

You can configure the following settings in the application with environment variables:

### Transcoding related configuration:

The project uses [Lumberjack](https://github.com/natefinch/lumberjack) for the log rotation of the ffmpeg transcoding processes.

#### RTSP_STREAM_CLEANUP_TIME
Default: `2m0s`<br/>
Type: string<br/>
Description: Time period for the cleanup process [info on format here](https://golang.org/pkg/time/#ParseDuration)<br/>

#### RTSP_STREAM_STORE_DIR
Default: `./videos`<br/>
Type: string<br/>
Description: Sub directory to store the video chunks<br/>

#### RTSP_STREAM_KEEP_FILES
Default: `false`<br/>
Type: bool<br/>
Description: Option to keep the chunks for the stream being transcoded<br/>

#### RTSP_STREAM_PROCESS_LOGGING_ENABLED
Default: `false`<br/>
Type: bool<br/>
Description: Indicates if logging of transcoding ffmpeg processes is enabled or not<br/>

#### RTSP_STREAM_PROCESS_LOGGING_DIR
Default: `/var/log/rtsp-stream`<br/>
Type: string<br/>
Description: Describes the directory where ffmpeg process logs are stored<br/>

#### RTSP_STREAM_PROCESS_LOGGING_MAX_SIZE
Default: `500`<br/>
Type: integer<br/>
Description: Maximum size of each log file in **megabytes** for retention<br/>

#### RTSP_STREAM_PROCESS_LOGGING_MAX_AGE
Default: `7`<br/>
Type: integer<br/>
Description: Maximum number of days that we store a given log file<br/>

#### RTSP_STREAM_PROCESS_LOGGING_MAX_BACKUPS
Default: `3`<br/>
Type: integer<br/>
Description: Maximum number of old log files to retain for each ffmpeg process<br/>

#### RTSP_STREAM_PROCESS_LOGGING_COMPRESS
Default: `true`<br/>
Type: bool<br/>
Description: Option to compress the rotated log file or not<br/>

<hr/>

### HTTP related configuration:

#### RTSP_STREAM_PORT
Default: `8080`<br/>
Type: integer<br/>
Description: Port where the application listens<br/>

#### RTSP_STREAM_DEBUG
Default: `false`<br/>
Type: bool<br/>
Description: Turns on / off debug features<br/>

#### RTSP_STREAM_LIST_ENDPOINT
Default: `false`<br/>
Type: bool<br/>
Description: Turns on / off the `/list` endpoint<br/>

<hr/>

### CORS related configuration

By default all origin is allowed to make requests to the server, but you might want to configure it for security reasons.

#### RTSP_STREAM_CORS_ENABLED
Default: `false`<br/>
Type: bool<br/>
Description: Indicates if cors should be handled as configured or as default (everything allowed)<br/>

#### RTSP_STREAM_CORS_ALLOWED_ORIGIN
Default: <br/>
Type: []string<br/>
Description: A list of origins a cross-domain request can be executed from<br/>

#### RTSP_STREAM_CORS_ALLOW_CREDENTIALS
Default: `false`<br/>
Type: bool<br/>
Description: Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates<br/>

#### RTSP_STREAM_CORS_MAX_AGE
Default: `0`<br/>
Type: integer<br/>
Description: Indicates how long (in seconds) the results of a preflight request can be cached<br/>

## UI

You can use the included UI for handling the streams. The UI is not a compact solution right now, but it gets the job done.

Running it with docker:

```s
docker run -p 80:80 -p 8080:8080 roverr/rtsp-stream:1-management
```

If you decide to use the management image, you should know that port 80 is flexible, you can set it to whatever you prefer, but 8080 is currently burnt into the UI as the ultimate port of the backend.

You should expect something like this:


<img src="./ui.gif"/>


## Debug

Debug information is described [here](docs/debugging/README.md)

## Proven players
The following list of players has been already tried out in production environment using this backend:

* Angular - [videogular](http://www.videogular.com/)
* React - [ReactHLS](https://github.com/foxford/react-hls)

## Coming soon
Codebase will be refactored as soon as I'll find some time to do it. 🙏
That will mean a major version bump. The goal is still the same. Keep it relatively simple and easy to integrate.
Solve the issue of not being able to play RTSP natively in browsers.

Plans for the future:
- Throw out URI based directory creation completely
- Add better logging and debug options
- Separate HTTP from Stream processing completely
- Add option to remove streams from the client (Could be tricky, gotta figure out if this should be an option even if non-authenticated mode is used)
- Add better documentation about how to debug streams
- Add documentation about how to create issues
- Add guide for PRs
