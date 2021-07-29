### rtsp流媒体服务器
    参考EasyDarwin编写的rtsp流媒体服务器,支持tcp传输,支持rtsp推拉流

### 部署运行
make run

### 测试
ffmpeg -re -i ./test2.mp4 -rtsp_transport tcp -vcodec h264 -f rtsp "rtsp://localhost:554/ChannelCode=1"  
ffplay -loglevel debug "rtsp://localhost:554/ChannelCode=1"

