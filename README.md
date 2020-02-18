# <img src="./docs/rtsp-stream.png"/>

[![Go Report Card](https://goreportcard.com/badge/github.com/Roverr/rtsp-stream)](https://goreportcard.com/report/github.com/Roverr/rtsp-stream)
 [![Maintainability](https://api.codeclimate.com/v1/badges/202152e83296250ab527/maintainability)](https://codeclimate.com/github/Roverr/rtsp-stream/maintainability)
 [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
 ![GitHub last commit](https://img.shields.io/github/last-commit/Roverr/rtsp-stream.svg)
 ![GitHub release](https://img.shields.io/github/release/Roverr/rtsp-stream.svg)

rtsp-stream is an easy to use, out of box solution that can be integrated into existing systems resolving the problem of not being able to play raw rtsp stream natively in browsers. 

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
* [Contributions and reporting issues](#contributions-and-reporting-issues)

## How does it work

The application converts raw `RTSP` streams into `HLS`.<br/>
The goal is make raw RTSP streams easily playable in browsers using HLS.

<p align="center">
  <img src="https://i.imgur.com/02X4uCX.png">
</p>

**Supports transcoding based on traffic**<br/>
The idea behind this is that it should not transcode anything until someone is actually watching the stream. This can help with network bottlenecks in systems where there are a lot of cameras installed.<br/>
There is a running go routine in the background that checks if a stream is being active or not. If it's not active anymore, the transcoding stops until the next request for that stream.

This functionality is configurable though so you can use it as a normal transcoding service if you would like.

## Run with Docker

Why should you use it with Docker?<br/>
Because the application relies on ffmpeg heavily therefore ensuring the environment is much easier with docker as everything comes with the image and you do not have to install anything besides docker. Other than installation, this way we can also avoid compatibility issues between operating systems.

The application has an offical [Docker repository](https://hub.docker.com/r/roverr/rtsp-stream/) at Dockerhub, therefore you can easily run it with simple commands:

```s
docker run -p 80:8080 roverr/rtsp-stream:2
```
## Easy API

There are 4 endpoints that are fully configurable to call
* `/start` - to start transcoding of a stream
* `/stream/id/*fileId` - static endpoint to serve video files for your browser
* `/list` - lists streams already known
* `/stop` - stops and removes a given stream

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

The service does not create any tokens, but your authentication service can create.<br/>
After it's created it can be validated in the transcoder using the same secret / keys.<br/>
It is the easiest way to integrate into existing systems.

<p align="center">
  <img width="600" height="500" src="https://i.imgur.com/j2dfmzf.png"/>
</p>

The following environment variables are available for this setup:

| Env variable | Description | Default | Type |
| :---        |    :----   |          ---: | :--- |
| RTSP_STREAM_AUTH_JWT_ENABLED | Indicates if the service should use the JWT authentication for the requests | `false` | bool |
| RTPS_STREAM_AUTH_JWT_SECRET | The secret used for creating the JWT tokens | `macilaci` | string |
| RTSP_STREAM_AUTH_JWT_PUB_PATH | Path to the public shared RSA key.| `/key.pub` | string |
| RTSP_STREAM_AUTH_JWT_METHOD | Can be `secret` or `rsa`. Changes how the application does the JWT verification.| `secret` | string |

You won't need the private key for it because no signing happens in this application.

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
docker run -p 80:80 -p 8080:8080 roverr/rtsp-stream:2-management
```

If you decide to use the management image, you should know that port 80 is flexible, you can set it to whatever you prefer, but 8080 is currently burnt into the UI as the ultimate port of the backend.

You should expect something like this:


<img src="./docs/ui.gif"/>


## Debug

Debug information is described [here](docs/debugging/README.md)

## Proven players
The following list of players has been already tried out in production environment using this backend:

* Angular - [videogular](http://www.videogular.com/)
* React - [ReactHLS](https://github.com/foxford/react-hls)

## Contributions and reporting issues

See more information about [this here](docs/contribution/README.md).
