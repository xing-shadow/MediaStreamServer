### rtsp流媒体服务器
    参考EasyDarwin
    rtsp流媒体服务器,tcp传输,支持rtsp推拉流

### rtmp流媒体服务器

```
参考livego
rtmp流媒体服务器,tcp传输,支持rtmp推拉流
```



### 部署运行

make run

### rtsp测试
ffmpeg -re -i ./test.mp4 -rtsp_transport tcp -c copy -f rtsp "rtsp://admin:123456@localhost:11554/ChannelCode=1"  
ffplay -loglevel debug "rtsp://admin:123456@localhost:11554/ChannelCode=1"

### rtmp测试

ffmpeg -re -i ./test.mp4 -c copy -f flv "rtmp://localhost:1935/live/test"
ffplay rtmp://localhost:1935/live/test

