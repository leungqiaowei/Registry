# Repository

## 1 简介

该项目是一个基于 Go 编程语言的“仓库”服务，提供高效的文件存储和查询功能。它支持文件的上传、下载、查询、删除和转发，具有可扩展的模块化设计，适用于分布式系统的存储需求。

## 2 功能

- **文件上传**：支持多种文件格式的高效上传。
- **文件下载**：通过 REST API 提供快速下载功能。
- **文件查询**：支持根据文件名和版本号的快速查询。
- **文件删除**：提供安全的文件删除功能。
- **文件转发**：支持多种文件格式高效转发到远程“仓库”（网络可达）。

## 3 文件结构

### 3.1 根目录

- `README.md`：项目的主要文档。
- `Makefile`：包含构建和运行任务的说明。
- `go.mod` 和 `go.sum`：Go 模块和依赖管理文件。

### 3.2 主要代码目录

#### 3.2.1 `cmd`

包含主程序入口文件：
- `cmd/registry/registry.go`：“仓库”服务的主入口。
- `cmd/registry/app/server.go`：服务端启动逻辑。

#### 3.2.2 `pkg`

包含项目的主要逻辑：
- `pkg/registry`：服务的实现，包括配置和工具函数。
  - `pkg/registry/server`：包含服务端处理逻辑，例如文件上传、下载和删除。
  - `pkg/registry/utils`：常用的工具函数，例如内存管理和文件操作。

#### 3.2.3`test`

- `test/registry/registry_test.go`：测试文件，包含对”仓库“服务功能的单元测试。

### 3.3 其他

- `docs/logs.md`：日志记录和管理文档。
- `docs/registry.md`：”仓库“功能和接口文档。
- `tmp/`：临时文件目录（用于存储日志和数据）。

## 4 API 接口

### 4.1 上传文件

**POST** `/upload`

- **使用 curl 测试**（`tag` 参数为可选项，默认为 `v1.0.0`；`isPermanent`参数为可选项，默认为`false`，如果设置为 `true`，在 `download` 则不会被删除）：

  ```shell
  curl -X POST -F "file=@example.txt" "http://<server_address>/upload?filename=<file>&tag=<version>&<isPermanent>=true"
  ```

### 4.2 下载文件

**GET** `/download`

- **使用 curl 测试**（`tag` 参数为可选项，默认为 `v1.0.0`）：

  ```shell
  curl -X GET "http://<server_address>/download?filename=<file>&tag=<version>"
  ```

### 4.3 查询文件

**GET** `/query/exits`

- **使用 curl 测试**（`tag` 参数为可选项，默认为 `v1.0.0`）：

  ```shell
  curl -X GET "http://<server_address>/query/exits?filename=<file>&tag=<version>"
  ```

### 4.4 删除文件

**DELETE** `/delete`

- **使用 curl 测试**（`tag` 参数为可选项，默认为 `v1.0.0`）：

  ```shell
  curl -X DELETE "http://<server_address>/delete?filename=<file>&tag=<version>"
  ```

### 4.5 转发文件

**POST** `/forward`

- **使用 curl 测试**（`tag` 参数为可选项，默认为 `v1.0.0`）：

  ```shell
  curl -X POST -F "file=@example.txt" "http://<server_address>/forward?filename=<file>&tag=<version>&target=http://<target_server_address>"
  ```

