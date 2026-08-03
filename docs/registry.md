## 1、项目简介
本项目主要提供数据的跨域转发服务，支持通过 Http 的方式传输转发数据

### 接口说明
项目主要提供三个接口：/get ；/post ；/post/forward；接收端口为 8081

#### 1、/get 接口接收 get 请求，用于获取仓库中存储的文件，需要传入文件名，

例如：curl http://localhost:8081/download?filename=test.txt 

#### 2、/post 接收 post 请求，接口用于上传文件，需要传入文件名和文件，
文件需要以二进制流的形式发送，需要设置 http 报文段的 Content-Type 
为 application/octet-stream 

例如 curl -X POST http://localhost:8081/upload?filename=test.txt 
--data-binary @test.txt -H "Content-Type: application/octet-stream"

#### 3、/post/forward 接口接收 post 请求，接口用于转发至指定目标仓库，需要传入目标仓库地址和文件名，
并且需要以二进制流的形式发送，需要设置 http 报文段的 Content-Type 为 application/octet-stream

例如： curl -X POST --data-binary @index.html 
-H "Content-Type: application/octet-stream" "http://localhost:8081/forward?target=http://10.244.169.151:8081&fileName=index.html"

其中 IP 地址需要自行替换；如需打包镜像部署，直接运行 Dockerfile 即可

仓库数据默认保存在 /tmp/data 目录下



 