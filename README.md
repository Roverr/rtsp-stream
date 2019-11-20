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
* [Easy API](#easy-api)
* [Authentication](#authentication)
    * [JWT](#jwt-authentication)
* [Configuration](#configuration)
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
## Easy API

There are 2 endpoint to call
* `/start` - to start transcoding of a stream
* `/stream/id/*fileId` - static endpoint to serve video files for your browser

[Read full documentation on API](docs/api/README.md).

## Authentication

The application offers different ways for authentication. There are situations when you can get away with no authentication, just
trusting requests because they are from reliable sources or just because they know how to use the API. In other cases, production cases, you definitely
want to protect the service. This application was not written to handle users and logins, so authentication is as lightweight as possible.

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

<img src="https://i.imgur.com/j2dfmzf.png"/>

## Configuration

The application tries to be as flexible as possible therefore there are a lot of configuration options available.
You can set the following information in the application:
* Sub directory where the application stores video chunks
* Time period for the cleanup process that stops streams if they are inactive
* Option to keep all video chunks forever instead of removing them when a stream becomes inactive
* Logging options for the underlying ffmpeg process
* CORS and other HTTP related options for the backend server itself
* Debug options for easier time when trying to find out what's wrong

Check the full list of environment variables [here](docs/configuration/README.md)

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

## Contributions and reporting issues

See more information about [this here](docs/contribution/README.md).

## Coming soon
Codebase will be refactored as soon as I'll find some time to do it. 🙏
That will mean a major version bump. The goal is still the same. Keep it relatively simple and easy to integrate.
Solve the issue of not being able to play RTSP natively in browsers.

Plans for the future:
- Add better logging and debug options
- Separate HTTP from Stream processing completely
- Add option to remove streams from the client (Could be tricky, gotta figure out if this should be an option even if non-authenticated mode is used)
- Add better documentation about how to debug streams
- Add documentation about how to create issues
- Add guide for PRs
