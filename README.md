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
* `RTSP_STREAM_PORT` - Port where the application listens
* `RTSP_STREAM_DEBUG` - Turns on debug logging

## Run with Docker
The application has an offical docker repository at dockerhub, therefore you can easily run it with simple commands:

`docker run -p 80:8080 roverr/rtsp-stream`

or you can build it yourself using the source code.


## Test it out
Create the following html file, then replace the source URL with your own choice.
```html
<?<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
</head>
<body>
        
<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
<video id="video"></video>
<script>
  var video = document.getElementById('video');
  if(Hls.isSupported()) {
    var hls = new Hls();
    hls.loadSource('http://localhost:8080/stream/host-name-here/index.m3u8');
    hls.attachMedia(video);
    hls.on(Hls.Events.MANIFEST_PARSED,function() {
      video.play();
  });
 }
 else if (video.canPlayType('application/vnd.apple.mpegurl')) {
    video.src = 'http://localhost:8080/stream/host-name-here/index.m3u8';
    video.addEventListener('loadedmetadata',function() {
      video.play();
    });
  }
</script>
</body>
</html>
```
