## Easy API
### POST /start

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
{ "uri": "/stream/id/index.m3u8" }
```

### GET /stream/id/*file

Simple static file serving which is used when fetching chunks of `HLS`. This will be called by the client (browser) to fetch the chunks of the stream based on the given `index.m3u8`.
Note that authentication will also be checked when accessing the files via this endpoint. Therefore for maximum performance you can turn off JWT authentication but it is not recommended at all.

### GET /list (debug)

This endpoint is used to list the streams in the system. 
Since the application does not handle users, it does not handle permissions obviously. 
You might not want everyone to be able to list the streams 
available in the system. But if you do, you can use this. You just have to enable it via [env variable](https://github.com/Roverr/rtsp-stream#configuration).



Response:
```js
[
    {
        "running": true,
        "uri": "/stream/9f4fa8eb-98c0-4ef6-9b89-b115d13bb192/index.m3u8"
    }
]
``` 
