# Registry 接口使用手册

本文档说明 Registry 服务当前提供的 HTTP 接口、云边反向隧道接口、请求参数、请求示例和响应格式。

## 1. 服务基础信息

默认服务地址：

```text
http://localhost:8119
```

如果服务部署在远端机器，请将示例中的 `localhost:8119` 替换为实际地址。

数据默认保存目录：

```text
./tmp/data
```

## 2. 命令行使用

### 2.1 启动云端服务

```bash
./registry
```

或直接运行：

```bash
go run ./cmd/registry
```

### 2.2 启动边侧 Agent

边侧 Agent 主动连接云端公网 IP，建立反向隧道：

```bash
./registry edge --cloud-url=ws://120.220.95.189:8119/edge/ws --edge-id=edge-a
```

所有参数：

| 参数 | 必填 | 默认值 | 说明 |
|---|---|---|---|
| `--cloud-url` | 是 | 无 | 云端公网 WebSocket 地址 |
| `--edge-id` | 否 | 主机名 | 边侧节点唯一标识 |
| `--local-base-url` | 否 | `http://127.0.0.1:8119` | 边侧本地要代理的服务 |
| `--token` | 否 | 无 | 云端鉴权 Token |
| `--reconnect-interval` | 否 | `5s` | 断线重连间隔 |
| `--heartbeat-interval` | 否 | `30s` | 心跳间隔 |
| `--request-timeout` | 否 | `25s` | 边侧请求本地服务超时 |

示例：边侧本地服务端口是 `8080`：

```bash
./registry edge \
  --cloud-url=ws://120.220.95.189:48119/edge/ws \
  --edge-id=factory-001 \
  --local-base-url=http://127.0.0.1:8080
```

### 2.3 查看帮助

```bash
./registry --help
./registry edge --help
```

---

## 3. 通用约定

### 3.1 默认版本 tag

所有涉及文件名和版本的接口中，如果未传 `tag`，默认使用：

```text
v1.0.0
```

### 3.2 文件类型 Header

涉及文件上传、接收、下载的接口支持请求头：

```http
FileType: file
```

可选值：

| 值 | 说明 |
|---|---|
| `file` | 普通文件，默认值 |
| `folder` | 文件夹 |
| `completion` | 文件夹传输完成标识，当前逻辑按文件夹处理 |

### 3.3 错误响应格式

多数接口错误返回为 JSON：

```json
{
  "error": "Bad Request",
  "message": "filename is required"
}
```

### 3.4 文件元数据格式

查询接口返回的文件元数据格式如下：

```json
{
  "file_name": "test.txt",
  "tag": "v1.0.0",
  "type": "file",
  "is_permanent": true,
  "file_path": "tmp/data/test.txt_v1.0.0",
  "storage_time": "2026-08-03T12:00:00+08:00",
  "owner": "127.0.0.1:12345",
  "size": 128,
  "file_hash": "sha256 hash"
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| `file_name` | 文件名 |
| `tag` | 文件版本 |
| `type` | 文件类型，`file` 或 `folder` |
| `is_permanent` | 是否永久保存 |
| `file_path` | 服务端实际存储路径 |
| `storage_time` | 存储时间 |
| `owner` | 上传方地址 |
| `size` | 文件大小，单位字节 |
| `file_hash` | 文件 SHA-256 哈希 |

---

# 4. 文件仓库接口

## 4.1 上传文件 `/upload`

### 功能说明

将文件上传到 Registry 本地仓库中。通过 `/upload` 上传的文件默认标记为永久文件，下载后不会自动删除。

### 请求方式

```http
POST /upload?filename=<文件名>&tag=<版本>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `filename` | 是 | 无 | 文件名 |
| `tag` | 否 | `v1.0.0` | 文件版本 |

### Header

| Header | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `FileType` | 否 | `file` | 文件类型 |
| `Content-Type` | 否 | 无 | 普通二进制上传建议使用 `application/octet-stream` |

### 请求示例

```bash
curl.exe -X POST "http://localhost:8119/upload?filename=test.txt&tag=v1.0.0" \
  -H "FileType: file" \
  -H "Content-Type: application/octet-stream" \
  --data-binary "@test.txt"
```

### 成功响应

```json
{
  "message": "file uploaded successfully"
}
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 未传 `filename` 或 `FileType` 非法 |
| `405` | 请求方法不是 `POST` |
| `500` | 文件保存失败，例如同名同版本文件已存在 |

---

## 4.2 下载文件 `/download`

### 功能说明

根据 `filename` 和 `tag` 下载文件。普通文件会直接返回二进制文件流。

如果文件是通过 `/receive` 接收并保存的临时文件，下载后会自动删除。

### 请求方式

```http
GET /download?filename=<文件名>&tag=<版本>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `filename` | 是 | 无 | 文件名 |
| `tag` | 否 | `v1.0.0` | 文件版本 |

### Header

| Header | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `FileType` | 否 | `file` | 文件类型 |

### 请求示例

```bash
curl.exe -X GET "http://localhost:8119/download?filename=test.txt&tag=v1.0.0" -o test.txt
```

### 成功响应

普通文件响应：

```http
HTTP/1.1 200 OK
Content-Type: application/octet-stream
```

响应体为文件内容。

文件夹类型响应：

```json
{
  "message": "folder dispatched successfully"
}
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 参数错误 |
| `404` | 文件不存在 |
| `405` | 请求方法不是 `GET` |
| `500` | 文件夹分发失败 |

---

## 4.3 接收文件 `/receive`

### 功能说明

接收其他节点发送来的文件。通过 `/receive` 接收的文件默认标记为临时文件，下载后会自动删除。

如果当前文件已经被订阅，则 `/receive` 收到文件后会优先转发给订阅者，而不是本地保存。

另外，当请求头 `FlowType: etcd` 时，该接口会作为特殊流量转发入口使用。

### 请求方式

```http
POST /receive?filename=<文件名>&tag=<版本>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `filename` | 普通文件接收时必填 | 无 | 文件名 |
| `tag` | 否 | `v1.0.0` | 文件版本 |
| `target` | `FlowType=etcd` 时必填 | 无 | etcd 流量目标 URL |

### Header

| Header | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `FileType` | 否 | `file` | 文件类型 |
| `FlowType` | 否 | 无 | 设置为 `etcd` 时走转发逻辑 |

### 请求示例：接收普通文件

```bash
curl.exe -X POST "http://localhost:8119/receive?filename=test.txt&tag=v1.0.0" \
  -H "FileType: file" \
  -H "Content-Type: application/octet-stream" \
  --data-binary "@test.txt"
```

### 请求示例：转发 etcd 流量

```bash
curl.exe -X POST "http://localhost:8119/receive?target=http://127.0.0.1:2379/v2/keys/test" \
  -H "FlowType: etcd" \
  --data-binary "@payload.bin"
```

### 成功响应

```json
{
  "message": "file received successfully"
}
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 参数错误或缺少 `target` |
| `405` | 请求方法不是 `POST` |
| `500` | 文件保存失败 |

---

## 4.4 删除文件 `/delete`

### 功能说明

根据 `filename` 和 `tag` 删除仓库中的文件或文件夹，并移除内存中的文件元数据。

### 请求方式

```http
DELETE /delete?filename=<文件名>&tag=<版本>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `filename` | 是 | 无 | 文件名 |
| `tag` | 否 | `v1.0.0` | 文件版本 |

### 请求示例

```bash
curl.exe -X DELETE "http://localhost:8119/delete?filename=test.txt&tag=v1.0.0"
```

### 成功响应

```json
{
  "message": "file test.txt (tag: v1.0.0) deleted successfully"
}
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 参数错误 |
| `404` | 文件不存在或删除失败 |
| `405` | 请求方法不是 `DELETE` |

---

## 4.5 查询单个文件 `/query/exits`

### 功能说明

查询指定 `filename` 和 `tag` 的文件元数据。

注意：当前接口路径为 `/query/exits`，这是代码中的现有拼写，不是 `/query/exists`。

### 请求方式

```http
GET /query/exits?filename=<文件名>&tag=<版本>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `filename` | 是 | 无 | 文件名 |
| `tag` | 否 | `v1.0.0` | 文件版本 |

### 请求示例

```bash
curl.exe "http://localhost:8119/query/exits?filename=test.txt&tag=v1.0.0"
```

### 成功响应

```json
{
  "file_name": "test.txt",
  "tag": "v1.0.0",
  "type": "file",
  "is_permanent": true,
  "file_path": "tmp/data/test.txt_v1.0.0",
  "storage_time": "2026-08-03T12:00:00+08:00",
  "owner": "127.0.0.1:12345",
  "size": 128,
  "file_hash": "sha256 hash"
}
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 参数错误 |
| `404` | 文件不存在 |
| `405` | 请求方法不是 `GET` |

---

## 4.6 查询文件列表 `/query/list`

### 功能说明

查询当前 Registry 内存索引中的全部文件元数据。

### 请求方式

```http
GET /query/list
```

### 请求示例

```bash
curl.exe "http://localhost:8119/query/list"
```

### 成功响应

```json
[
  {
    "file_name": "test.txt",
    "tag": "v1.0.0",
    "type": "file",
    "is_permanent": true,
    "file_path": "tmp/data/test.txt_v1.0.0",
    "storage_time": "2026-08-03T12:00:00+08:00",
    "owner": "127.0.0.1:12345",
    "size": 128,
    "file_hash": "sha256 hash"
  }
]
```

如果没有文件，返回：

```json
[]
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `405` | 请求方法不是 `GET` |

---

## 4.7 转发文件 `/forward`

### 功能说明

将当前请求转发到另一个节点的 `/receive` 接口。该接口用于节点之间转发数据。

服务会根据请求头中的 `ClusterID` 构造目标地址，并将路径中的 `/forward` 替换为 `/receive`。

### 请求方式

```http
POST /forward?filename=<文件名>&tag=<版本>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `filename` | 是 | 无 | 文件名 |
| `tag` | 否 | `v1.0.0` | 文件版本 |

### Header

| Header | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `ClusterID` | 是 | 无 | 目标节点 IP 或主机名 |
| `FileType` | 否 | `file` | 文件类型 |

### 请求示例

```bash
curl.exe -X POST "http://localhost:8119/forward?filename=test.txt&tag=v1.0.0" \
  -H "ClusterID: 10.0.0.12" \
  -H "FileType: file" \
  -H "Content-Type: application/octet-stream" \
  --data-binary "@test.txt"
```

如果当前请求访问的是：

```text
http://localhost:8119/forward?filename=test.txt&tag=v1.0.0
```

并且请求头为：

```http
ClusterID: 10.0.0.12
```

则转发目标大致为：

```text
http://10.0.0.12:8119/receive?filename=test.txt&tag=v1.0.0
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 缺少 `ClusterID` |
| `502` | 目标节点不可达或目标节点返回错误 |

---

# 5. 文件订阅接口

## 5.1 订阅文件 `/subscribe`

### 功能说明

注册对某个文件版本的订阅关系。后续当 `/receive` 收到同名同版本文件时，会将请求转发给订阅者。

订阅目标地址格式为：

```text
http://<client_ip>:8080/receive?filename=<filename>&tag=<tag>
```

### 请求方式

```http
POST /subscribe?filename=<文件名>&tag=<版本>&client_ip=<客户端IP>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `filename` | 是 | 无 | 订阅的文件名 |
| `tag` | 否 | `v1.0.0` | 文件版本 |
| `client_ip` | 是 | 无 | 订阅者 IP |

### 请求示例

```bash
curl.exe -X POST "http://localhost:8119/subscribe?filename=test.txt&tag=v1.0.0&client_ip=192.168.1.20"
```

### 成功响应

```json
{
  "message": "subscription successful"
}
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 缺少 `filename` 或 `client_ip` |
| `405` | 请求方法不是 `POST` |

---

## 5.2 查询订阅列表 `/subscribe/list`

### 功能说明

查询当前所有订阅关系。

### 请求方式

```http
GET /subscribe/list
```

### 请求示例

```bash
curl.exe "http://localhost:8119/subscribe/list"
```

### 成功响应

```json
[
  [
    "test.txt_v1.0.0",
    "http://192.168.1.20:8080/receive?filename=test.txt&tag=v1.0.0"
  ]
]
```

如果没有订阅，返回：

```json
[]
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `405` | 请求方法不是 `GET` |

---

# 6. 云边反向隧道接口

云边反向隧道用于解决以下场景：

```text
云端有公网 IP，边侧在内网或 NAT 后面。
公网无法主动连接边侧，但边侧可以主动访问云端公网 IP。
```

设计方式：

```text
边侧 Agent 主动连接云端 /edge/ws。
云端维护边侧连接池。
云端收到 /edges/{edge_id}/... 请求后，通过 WebSocket 隧道转发到边侧本地服务。
边侧执行本地 HTTP 请求后，将响应通过 WebSocket 返回云端。
```

## 6.1 边侧建立长连接 `/edge/ws`

### 功能说明

边侧主动向云端发起 WebSocket 长连接，建立云边反向隧道。

该接口不是普通 HTTP 接口，而是 WebSocket 接口。

### 请求方式

```http
GET /edge/ws?edge_id=<边侧节点ID>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `edge_id` | 是 | 无 | 边侧节点唯一标识 |

### Header

| Header | 必填 | 说明 |
|---|---:|---|
| `X-Edge-ID` | 否 | 可作为 `edge_id` 的备用来源 |
| `Authorization` | 否 | 预留鉴权字段 |

### 连接示例

边侧连接云端：

```text
ws://120.220.95.189:48119/edge/ws?edge_id=edge-a
```

如果本地测试云端服务：

```text
ws://localhost:8119/edge/ws?edge_id=edge-a
```

### 消息协议

基础消息格式：

```json
{
  "type": "heartbeat",
  "id": "req-xxx",
  "edge_id": "edge-a",
  "timestamp": "2026-08-03T12:00:00+08:00",
  "payload": {}
}
```

消息类型：

| 类型 | 方向 | 说明 |
|---|---|---|
| `register` | Edge -> Cloud | 注册边侧节点 |
| `heartbeat` | Edge -> Cloud | 心跳 |
| `request` | Cloud -> Edge | 云端请求边侧 |
| `response` | Edge -> Cloud | 边侧响应云端 |
| `error` | 双向 | 错误消息 |

---

## 6.2 查询在线边侧节点 `/edge/clients`

### 功能说明

查询当前已经连接到云端的边侧节点列表。

### 请求方式

```http
GET /edge/clients
```

### 请求示例

```bash
curl.exe "http://localhost:8119/edge/clients"
```

### 成功响应

```json
[
  {
    "edge_id": "edge-a",
    "remote_addr": "10.0.0.2:53212",
    "connected_at": "2026-08-03T12:00:00+08:00",
    "last_seen": "2026-08-03T12:00:30+08:00",
    "status": "online"
  }
]
```

如果没有边侧节点在线，返回：

```json
[]
```

---

## 6.3 检查边侧健康 `/edge/health`

### 功能说明

云端通过 WebSocket 隧道请求指定边侧节点的本地 `/health` 接口。

### 请求方式

```http
GET /edge/health?edge_id=<边侧节点ID>
```

### Query 参数

| 参数 | 必填 | 默认值 | 说明 |
|---|---:|---|---|
| `edge_id` | 是 | 无 | 要检查的边侧节点 ID |

### 请求示例

```bash
curl.exe "http://localhost:8119/edge/health?edge_id=edge-a"
```

### 成功响应

响应内容取决于边侧本地服务 `/health` 的返回。

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 缺少 `edge_id` |
| `502` | 边侧不在线、隧道请求超时或边侧本地服务不可达 |

---

## 6.4 云端代理访问边侧 `/edges/{edge_id}/{target_path}`

### 功能说明

云端通过 WebSocket 隧道代理访问边侧内网服务。

这是云边反向隧道的核心接口。

### 请求方式

```http
ANY /edges/{edge_id}/{target_path}
```

其中：

| 路径段 | 说明 |
|---|---|
| `{edge_id}` | 边侧节点 ID |
| `{target_path}` | 要在边侧本地访问的路径 |

边侧 Agent 默认会将该请求转发到：

```text
http://127.0.0.1:8119/{target_path}
```

如果边侧 Agent 配置了其他 `local_base_url`，则以配置为准。

### 示例：云端查询边侧文件列表

云端请求：

```bash
curl.exe "http://localhost:8119/edges/edge-a/query/list"
```

边侧实际执行：

```http
GET http://127.0.0.1:8119/query/list
```

### 示例：云端上传文件到边侧

云端请求：

```bash
curl.exe -X POST "http://localhost:8119/edges/edge-a/upload?filename=test.txt&tag=v1.0.0" \
  -H "FileType: file" \
  -H "Content-Type: application/octet-stream" \
  --data-binary "@test.txt"
```

边侧实际执行：

```http
POST http://127.0.0.1:8119/upload?filename=test.txt&tag=v1.0.0
```

### 示例：云端下载边侧文件

```bash
curl.exe -X GET "http://localhost:8119/edges/edge-a/download?filename=test.txt&tag=v1.0.0" -o test.txt
```

边侧实际执行：

```http
GET http://127.0.0.1:8119/download?filename=test.txt&tag=v1.0.0
```

### 示例：云端删除边侧文件

```bash
curl.exe -X DELETE "http://localhost:8119/edges/edge-a/delete?filename=test.txt&tag=v1.0.0"
```

边侧实际执行：

```http
DELETE http://127.0.0.1:8119/delete?filename=test.txt&tag=v1.0.0
```

### 常见错误

| 状态码 | 场景 |
|---:|---|
| `400` | 路径格式错误，不符合 `/edges/{edge_id}/{target_path}` |
| `502` | 边侧节点不在线、隧道超时或边侧本地服务请求失败 |

---

# 7. 云边隧道协议详情

## 7.1 TunnelMessage

云端和边侧通过 WebSocket 传输统一消息：

```json
{
  "type": "request",
  "id": "req-123",
  "edge_id": "edge-a",
  "timestamp": "2026-08-03T12:00:00+08:00",
  "payload": {}
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| `type` | 消息类型 |
| `id` | 请求 ID，用于匹配请求和响应 |
| `edge_id` | 边侧节点 ID |
| `timestamp` | 消息时间 |
| `payload` | 具体消息体 |

## 7.2 request payload

云端发给边侧的请求消息：

```json
{
  "method": "GET",
  "path": "/query/list",
  "query": {},
  "header": {},
  "body": null
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| `method` | HTTP 方法 |
| `path` | 边侧本地服务路径 |
| `query` | Query 参数 |
| `header` | HTTP Header |
| `body` | 请求体，二进制内容会经过 JSON 编码传输 |

## 7.3 response payload

边侧返回给云端的响应消息：

```json
{
  "status_code": 200,
  "header": {
    "Content-Type": ["application/json"]
  },
  "body": "...",
  "error": ""
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| `status_code` | 边侧本地服务响应状态码 |
| `header` | 边侧本地服务响应 Header |
| `body` | 边侧本地服务响应体 |
| `error` | 错误信息，成功时为空 |

---

# 8. 推荐测试流程

## 8.1 测试普通文件接口

```bash
curl.exe -X POST "http://localhost:8119/upload?filename=test.txt&tag=v1" --data-binary "@test.txt"
curl.exe "http://localhost:8119/query/exits?filename=test.txt&tag=v1"
curl.exe "http://localhost:8119/download?filename=test.txt&tag=v1" -o test-download.txt
curl.exe -X DELETE "http://localhost:8119/delete?filename=test.txt&tag=v1"
curl.exe "http://localhost:8119/query/list"
```

## 8.2 测试边缘基础接口

```bash
curl.exe "http://localhost:8119/edge/clients"
curl.exe "http://localhost:8119/edge/health?edge_id=edge-a"
curl.exe "http://localhost:8119/edges/edge-a/query/list"
```

如果 `edge-a` 未连接，预期返回 `502`。

## 8.3 测试完整云边链路

完整链路需要先启动云端服务和边侧 Agent：

**步骤 1：云端启动 Registry**

```bash
./registry
```

**步骤 2：边侧连接云端**

```bash
./registry edge --cloud-url=ws://<cloud-public-ip>:<port>/edge/ws --edge-id=edge-a
```

**步骤 3：云端访问边侧**

```bash
curl.exe "http://<cloud-public-ip>:<port>/edges/edge-a/query/list"
```

如果返回边侧的文件列表，说明云端已通过反向隧道成功访问边侧内网服务。

同样可以操作边侧的其他接口：

```bash
curl.exe "http://<cloud-public-ip>:<port>/edges/edge-a/upload?filename=test.txt" --data-binary "@test.txt"
curl.exe "http://<cloud-public-ip>:<port>/edges/edge-a/download?filename=test.txt&tag=v1.0.0" -o test.txt
```

---

# 9. 接口总表

| 分类 | 接口 | 方法 | 说明 |
|---|---|---|---|
| 文件仓库 | `/upload` | `POST` | 上传并永久保存文件 |
| 文件仓库 | `/download` | `GET` | 下载文件 |
| 文件仓库 | `/receive` | `POST` | 接收文件或转发 etcd 流量 |
| 文件仓库 | `/delete` | `DELETE` | 删除文件 |
| 文件仓库 | `/query/exits` | `GET` | 查询单个文件元数据 |
| 文件仓库 | `/query/list` | `GET` | 查询文件列表 |
| 文件仓库 | `/forward` | `POST` | 转发请求到其他节点 `/receive` |
| 订阅 | `/subscribe` | `POST` | 订阅文件 |
| 订阅 | `/subscribe/list` | `GET` | 查询订阅列表 |
| 云边隧道 | `/edge/ws` | `GET WebSocket` | 边侧主动连接云端 |
| 云边隧道 | `/edge/clients` | `GET` | 查询在线边侧节点 |
| 云边隧道 | `/edge/health` | `GET` | 检查指定边侧健康 |
| 云边隧道 | `/edges/{edge_id}/{target_path}` | 任意 | 云端代理访问边侧内网服务 |
