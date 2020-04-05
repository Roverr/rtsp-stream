## Easy API

Endpoints are [configured](#configuration) using `rtsp-stream.yml` which is an easy way to define permissions for each endpoint.

* [/start](#post-start) - Starts transcoding for a given raw rtsp stream
* [/stream/{id}](#get-streamidfile) - Static file serving for getting HLS video chunks
* [/list](#get-list) - Lists available streams
* [/stop](#post-stop) - Stops transcoding for given stream without removing it

### Configuration

Application can be configured using an `rtsp-stream.yml` put next to the binary file.
Looks like the following:
```yaml
version: 1.0
endpoints:
  start:
    enabled: true
    secret: macilaci
  static:
    enabled: true
    secret: macilaci
  stop:
    enabled: true
    secret: macilaci
  list:
    enabled: true
    secret: macilaci
listen:
   - alias: camera1
     uri: rtp://user:pass@host/camera/123
     enabled: false
```
* enabled - false by default - boolean that indicates if the given endpoint is enabled or not
* secret - empty by default - string which will be the secret in the JWT token

The application will decode the JWT token used for authentication and look for the given secret value in the token. If the secret matches the request will be successful.

If the secret is left empty, it won't require any authentication to reach the given endpoint. Same secret values can be used to create tiered list for users.

```yaml
version: 1.0
endpoints:
  start:
    enabled: true
  static:
    enabled: true
  stop:
    enabled: true
    secret: macilaci
  list:
    enabled: true
    secret: macilaci
listen:
   - alias: camera1
     uri: rtp://user:pass@host/camera/123
     enabled: false
```

In this example everyone can start a stream and fetch video chunks if they know the id of the video, but only users with macilaci secret will be able to access the stop and list endpoints. 

This behaviour is changed when JWT authentication is enabled. In that case everyone will have to have a valid token, but only the given endpoints with secret value will be checked fro secret.

If you are using **Docker** you can add your local file in the following way:
```s
docker run -v `pwd`/rtsp-stream.yml:/app/rtsp-stream.yml \
           -p 8080:8080 \
           -e RTSP_STREAM_DEBUG=true \
           roverr/rtsp-stream:2
``` 

More commands around docker at [debugging](../debugging#Docker)

```yaml
listen:
   - alias: camera1
     uri: rtp://user:pass@host/camera/123
     enabled: false
```

**listen** is used for preloading streams into the system. While ID generation here is not possible, the introduction of aliases helps in overcoming this issue. Listen is an array of streams to preload into the system.
* alias - Used as the reference when starting a stream
* uri - The URI for the camera source
* enabled - Indicates if the system should load the given record or not

### POST /start

Starts the transcoding of the given stream. You have to pass URI format with rtsp procotol. 
The respond should be considered the subpath for the video player to call.
So if your applicaiton is `myapp.com` then you should call `myapp.com/stream/host/index.m3u8` in your video player.
The reason for this is to remain flexible regarding useability. The alias value is optional.

Requires payload:
```js
{ 
    "uri": "rtsp://username:password@host",
    "alias": "camera1"
}
```

Response:
```js
{ 
    "uri": "/stream/id/index.m3u8",
    "running": true,
    "id": "id",
    "alias": "camera1"
}
```

**alias** is now available as a secondary reference for the stream. This means that you can reference a stream, by using its alias instead of its ID.<br/>
Existing aliases can be overwritten. As the API is still URI based, the best case for using them is when preloading a stream.

### GET /stream/{id}/*file

Simple static file serving which is used when fetching chunks of `HLS`. This will be called by the client (browser) to fetch the chunks of the stream based on the given `index.m3u8`.
Note that authentication will also be checked when accessing the files via this endpoint. Therefore for maximum performance you can turn off JWT authentication but it is not recommended at all.
The id value can either be the uuid of the stream or the alias if available.

### GET /list

This endpoint is used to list the streams in the system.

Response:
```js
[
    {
        "running": true,
        "uri": "/stream/9f4fa8eb-98c0-4ef6-9b89-b115d13bb192/index.m3u8",
        "id": "9f4fa8eb-98c0-4ef6-9b89-b115d13bb192",
        "alias": "9f4fa8eb-98c0-4ef6-9b89-b115d13bb192"
    },
    {
        "running": false,
        "uri": "/stream/camera1/index.m3u8",
        "id": "8ab9a89c-8271-4c89-97b7-c91372f4c1b0",
        "alias": "camera1"
    }
]
``` 

### POST /stop

Endpoint used for stopping and removing a stream from the stored list. Either include an ID or Alias value to identify the stream. 
In the event both values are provided, the ID will be used.

Requires payload:
```js
{ 
    "id": "40b1cc1b-bf19-4b07-8359-e934e7222109",
    "alias": "camera1",
    "remove": true, // optional - indicates if stream should be removed as well from list or not
    "wait": false // optional - indicates if the call should wait for the stream to stop
}
```

Response:
Empty 200
Empty 404
