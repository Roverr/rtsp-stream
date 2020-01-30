# Debugging rtsp-stream processing

## Table of contents
* [General](#general)
* [Docker](#docker)
    * [Backend logs](#backend-logs)
    * [Management logs](#management-logs)

## General

From v2.0.0 the service makes it easier for you to debug. By setting `RTSP_STREAM_DEBUG` env variable enabled the following configuration will be overwritten:
- **RTSP_STREAM_KEEP_FILES** - true - No video will be deleted during debugging
- **RTSP_STREAM_PROCESS_LOGGING** - true - Logs will be available under `/var/log/rtsp-stream/` for ffmpeg processes
- **RTSP_STREAM_AUDIO_ENABLED** - false - Won't transcode any audio

[VLC](https://www.videolan.org/vlc/) is recommended for validiting transcoding issues as it plays everything.

## Docker
### Backend logs

Obtaining logs for your failed process is the number one priority when you are debugging.
After setting the debug environment variable the service is now logging processes as well.

When the client call the `/start` endpoint with an URI the service will try to use that and put into an ffmpeg process. 
This will be logged on the console:

```s
Imres-MacBook-Pro:rtsp-stream rover$ docker run -e RTSP_STREAM_DEBUG=true -p 8080:8080 roverr/rtsp-stream:2
time="2019-11-18T17:35:47Z" level=info msg="RTSP-STREAM started on 8080"
time="2019-11-18T17:36:17Z" level=info msg="rtsp://admin:password123@hosting.dyndns.org:554/Streaming/Channels/102 started processing"
time="2019-11-18T17:36:17Z" level=debug msg="Created stream with storepath ./videos/784bf8eb-6082-43fe-a98b-72a6ddd6c02f"
```

You can recognise that the loaded video will be ID'd as `784bf8eb-6082-43fe-a98b-72a6ddd6c02f`. In this case you can find the logs of the underlying ffmpeg process in `/var/log/rtsp-stream/784bf8eb-6082-43fe-a98b-72a6ddd6c02f.log`

```s
Imres-MacBook-Pro:rtsp-stream rover$ docker ps
CONTAINER ID        IMAGE                  COMMAND                  CREATED             STATUS              PORTS                      NAMES
4527f8008fe3        roverr/rtsp-stream:2   "/app/server"            2 minutes ago       Up 2 minutes        0.0.0.0:8080->8080/tcp     youthful_ride

Imres-MacBook-Pro:rtsp-stream rover$ docker exec -it 4527f8008fe3 /bin/bash
bash-4.4# cat /var/log/rtsp-stream/784bf8eb-6082-43fe-a98b-72a6ddd6c02f.log 
```

Ideally you should see something like this:
```s
ffmpeg version 3.4 Copyright (c) 2000-2017 the FFmpeg developers
  built with gcc 6.4.0 (Alpine 6.4.0)
  configuration: --prefix=/usr --enable-avresample --enable-avfilter --enable-gnutls --enable-gpl --enable-libmp3lame --enable-librtmp --enable-libvorbis --enable-libvpx --enable-libxvid --enable-libx264 --enable-libx265 --enable-libtheora --enable-libv4l2 --enable-postproc --enable-pic --enable-pthreads --enable-shared --enable-libxcb --disable-stripping --disable-static --enable-vaapi --enable-vdpau --enable-libopus --disable-debug
  libavutil      55. 78.100 / 55. 78.100
  libavcodec     57.107.100 / 57.107.100
  libavformat    57. 83.100 / 57. 83.100
  libavdevice    57. 10.100 / 57. 10.100
  libavfilter     6.107.100 /  6.107.100
  libavresample   3.  7.  0 /  3.  7.  0
  libswscale      4.  8.100 /  4.  8.100
  libswresample   2.  9.100 /  2.  9.100
  libpostproc    54.  7.100 / 54.  7.100
Input #0, rtsp, from 'rtsp://admin:password123@host.dyndns.org:554/Streaming/Channels/102':
  Metadata:
    title           : HIK Media Server V3.4.103
    comment         : HIK Media Server Session Description : standard
  Duration: N/A, start: 0.401000, bitrate: N/A
    Stream #0:0: Video: h264 (Main), yuv420p(progressive), 352x288, 25 fps, 25 tbr, 90k tbn, 50 tbc
[hls @ 0x55b50eff49c0] Opening './videos/784bf8eb-6082-43fe-a98b-72a6ddd6c02f/0.ts' for writing
Output #0, hls, to './videos/784bf8eb-6082-43fe-a98b-72a6ddd6c02f/index.m3u8':
  Metadata:
    title           : HIK Media Server V3.4.103
    comment         : HIK Media Server Session Description : standard
    encoder         : Lavf57.83.100
    Stream #0:0: Video: h264 (Main), yuv420p(progressive), 352x288, q=2-31, 25 fps, 25 tbr, 90k tbn, 25 tbc
Stream mapping:
  Stream #0:0 -> #0:0 (copy)
Press [q] to stop, [?] for help
[hls @ 0x55b50eff49c0] Opening './videos/784bf8eb-6082-43fe-a98b-72a6ddd6c02f/1.ts' for writing
[hls @ 0x55b50eff49c0] Opening './videos/784bf8eb-6082-43fe-a98b-72a6ddd6c02f/index.m3u8.tmp' for writing
[hls @ 0x55b50eff49c0] Opening './videos/784bf8eb-6082-43fe-a98b-72a6ddd6c02f/2.ts' for writing
[hls @ 0x55b50eff49c0] Opening './videos/784bf8eb-6082-43fe-a98b-72a6ddd6c02f/index.m3u8.tmp' for writing
[hls @ 0x55b50eff49c0] Opening './videos/784bf8eb-6082-43fe-a98b-72a6ddd6c02f/3.ts' for writing
```

But for example when you are trying to add something that is not streaming right now:
```s
time="2019-11-18T17:45:32Z" level=info msg="rtsp://nonexistent:invalid@hosting.dyndns.org:554/Streaming/Channels/102 started processing"
time="2019-11-18T17:45:32Z" level=debug msg="Created stream with storepath ./videos/de7a74b9-1d0e-4bbc-b815-8820fce52186"
time="2019-11-18T17:45:32Z" level=error msg="Error happened during starting of ./videos/de7a74b9-1d0e-4bbc-b815-8820fce52186/index.m3u8 || Error: exit status 1"
```

```s
bash-4.4# cat /var/log/rtsp-stream/de7a74b9-1d0e-4bbc-b815-8820fce52186.log 
ffmpeg version 3.4 Copyright (c) 2000-2017 the FFmpeg developers
  built with gcc 6.4.0 (Alpine 6.4.0)
  configuration: --prefix=/usr --enable-avresample --enable-avfilter --enable-gnutls --enable-gpl --enable-libmp3lame --enable-librtmp --enable-libvorbis --enable-libvpx --enable-libxvid --enable-libx264 --enable-libx265 --enable-libtheora --enable-libv4l2 --enable-postproc --enable-pic --enable-pthreads --enable-shared --enable-libxcb --disable-stripping --disable-static --enable-vaapi --enable-vdpau --enable-libopus --disable-debug
  libavutil      55. 78.100 / 55. 78.100
  libavcodec     57.107.100 / 57.107.100
  libavformat    57. 83.100 / 57. 83.100
  libavdevice    57. 10.100 / 57. 10.100
  libavfilter     6.107.100 /  6.107.100
  libavresample   3.  7.  0 /  3.  7.  0
  libswscale      4.  8.100 /  4.  8.100
  libswresample   2.  9.100 /  2.  9.100
  libpostproc    54.  7.100 / 54.  7.100
[rtsp @ 0x7f7a6ea35600] method OPTIONS failed: 401 Unauthorized
rtsp://nonexistent:invalid@hosting.dyndns.org:554/Streaming/Channels/102: Server returned 401 Unauthorized (authorization failed)
```

While `exit status 1` is not the most detailed error, this is because the console of the service should not reflect underlying errors as they are. As the service is an abstraction over ffmpeg processes this is a much clearer way to obtain error messages.

Furthermore you can attach a volume to the container to collect logs locally to a logs directory:
```s
docker run -v `pwd`/logs:/var/log/ -p 8080:8080 -e RTSP_STREAM_DEBUG=true roverr/rtsp-stream:2
```
Now you can do debugging without even going into the container.

### Management logs

Obtaining logs in the management image is a bit trickier. As the management line is not the most supported way of using this service, you can encounter issues with it as well. Currently a systemd handles the processes which use the following configuration:
```s
[supervisord]
logfile = /tmp/supervisord.log
[program:rtsp-stream]
command=/app/server
autostart=true
autorestart=true
stderr_logfile=/var/log/rtsp-stream.err.log
stdout_logfile=/var/log/rtsp-stream.out.log
environment=RTSP_STREAM_DEBUG=true
[program:rtsp-stream-ui]
command=http-server -p 80 /ui/
autostart=true
autorestart=true
stderr_logfile=/var/log/rtsp-stream-ui.err.log
stdout_logfile=/var/log/rtsp-stream-ui.out.log
environment=API_URL=http://127.0.0.1:8080
```
Which means that when you are running the management image you will see the direct logs of the `supervisord`. This is needed because image has to run the transcoder backend and a frontend serving as well.
With the UI solution the following files are created in `/var/log`: 
* rtsp-stream-ui.err.log
* rtsp-stream-ui.out.log
* rtsp-stream.err.log
* rtsp-stream.out.log

Usually you won't see anything in rtsp-stream.err.log and nothing in general in the ui logs. (As it is just a really simple http server)


The same rule works here as it works in the simple service. Setting `RTSP_STREAM_DEBUG=true` will enable process logging. Therefore logs for the processess will be also created in this image under `/var/log/rtsp-stream/`

```s
docker run -v `pwd`/logs:/var/log/ -p 80:80 -p 8080:8080 roverr/rtsp-stream:2-management
```
