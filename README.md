# 🚀 gRPC-ToDoList 分布式文件存储系统

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![gRPC](https://img.shields.io/badge/gRPC-1.50+-green.svg)](https://grpc.io)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

本项目是基于 Go 的**企业级分布式文件存储系统**，集成了**备忘录管理**和**云存储网盘**功能。支持**表单上传 + 流式上传 + 异步处理 + 秒传 + 全盘搜索**，采用微服务架构设计，支持通过 Kafka 进行异步任务调度，并使用七牛云作为云存储解决方案。

## ✨ 核心特性

- 🔐 **用户认证系统** - JWT Token 认证，支持用户注册、登录、权限管理
- 📝 **备忘录管理** - 支持 CRUD 操作，任务分类，状态管理
- 📁 **多种上传方式** - 表单上传、流式上传、异步上传
- ⚡ **智能秒传** - 基于 SHA256 哈希的跨用户秒传机制
- 🔍 **全盘搜索** - 支持文件名模糊搜索，跨用户文件发现
- ☁️ **云存储集成** - 七牛云对象存储，CDN 加速访问
- 🏗️ **微服务架构** - Gateway、User、Task、Files 四大服务模块
- 📊 **异步处理** - Kafka 消息队列，提升系统响应性能
- 🔄 **服务发现** - etcd 注册中心，支持服务自动发现与负载均衡

## 🏗️ 系统架构

### 微服务模块

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Gateway       │    │     User        │    │     Task        │    │     Files       │
│   (端口: 4000)   │    │  (端口: 10002)   │    │  (端口: 10003)   │    │  (端口: 10004)   │
│                 │    │                 │    │                 │    │                 │
│ • HTTP API      │    │ • 用户注册登录    │    │ • 备忘录管理     │    │ • 文件上传下载   │
│ • 用户认证      │    │ • JWT Token     │    │ • 任务分类      │    │ • 秒传检测      │
│ • 请求路由      │    │ • 权限管理      │    │ • 状态管理      │    │ • 云存储集成    │
│ • 负载均衡      │    │                 │    │                 │    │ • 全盘搜索      │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │                       │
         └───────────────────────┼───────────────────────┼───────────────────────┘
                                 │                       │
                    ┌─────────────────────────────────────────────────────┐
                    │                  基础设施层                          │
                    │                                                     │
                    │  etcd (服务发现)  │  MySQL (数据存储)  │  Redis (缓存)  │
                    │  Kafka (消息队列) │  七牛云 (对象存储)                   │
                    └─────────────────────────────────────────────────────┘
```

### 技术栈详情

#### 🔧 后端技术

| 层级       | 技术            | 版本   | 描述                                                 |
| ---------- | --------------- | ------ | ---------------------------------------------------- |
| Web 框架   | Gin             | v1.9+  | 轻量级 HTTP Web 框架，负责 API 网关和用户认证        |
| RPC 通信   | gRPC + Protobuf | v1.50+ | 高性能服务间通信，支持流式大文件上传                  |
| 数据库 ORM | GORM            | v1.25+ | Go 语言主流 ORM 框架，简化数据库操作                 |
| 日志系统   | Zap             | v1.24+ | 高性能结构化日志库，支持链路追踪                     |
| 配置管理   | Viper           | v1.15+ | 灵活的配置文件管理，支持多种格式                     |

#### 🗄️ 存储与中间件

| 组件        | 版本    | 功能     | 说明                                       |
| ----------- | ------- | -------- | ------------------------------------------ |
| **MySQL**   | 8.0+    | 数据库   | 存储用户信息、文件元数据、备忘录等         |
| **Redis**   | 7.0+    | 缓存     | 用户会话缓存、秒传哈希缓存                 |
| **Kafka**   | 2.8+    | 消息队列 | 异步文件处理、任务调度（支持重试机制）     |
| **etcd**    | 3.5+    | 注册中心 | 服务注册发现、配置管理、分布式锁           |
| **七牛云**  | SDK v7  | 对象存储 | 文件云存储、CDN 加速、跨区域访问           |

## 📁 功能模块

### 🔐 用户系统
- **用户注册/登录** - 支持用户名密码认证
- **JWT Token 认证** - 无状态认证，支持 Token 刷新
- **权限管理** - 基于用户 ID 的资源访问控制

### 📝 备忘录系统
- **CRUD 操作** - 创建、读取、更新、删除备忘录
- **分类管理** - 支持自定义分类标签
- **状态跟踪** - 待办、进行中、已完成状态管理

### 📂 文件存储系统

#### 上传方式
- ✅ **表单上传** - 适用于小文件（< 10MB），支持批量上传
- ✅ **流式上传** - 适用于大文件，1MB 分片，支持断点续传
- ✅ **异步上传** - Kafka 队列处理，提升响应速度
- ✅ **智能秒传** - SHA256 哈希检测，跨用户文件共享

#### 存储特性
- 🔍 **全盘搜索** - 文件名模糊搜索，支持分页和过滤
- 📥 **跨用户下载** - 支持下载其他用户的公开文件
- 🗑️ **智能删除** - 安全删除机制，保护共享文件
- ☁️ **云存储集成** - 七牛云对象存储，全球 CDN 加速

## 🚀 快速开始

### 环境要求

- **Go**: 1.19+
- **Docker**: 20.10+
- **Docker Compose**: 2.0+

### 启动基础设施

使用 Docker Compose 一键启动所有依赖服务：

```bash
# 启动所有基础设施服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

启动的服务包括：
- **MySQL** (端口: 3306) - 主数据库
- **Redis** (端口: 6379) - 缓存服务
- **etcd** (端口: 2379) - 服务注册中心
- **Kafka** (端口: 9092) - 消息队列
- **Zookeeper** (端口: 2181) - Kafka 协调服务

### 配置七牛云

复制配置模板并填入你的七牛云信息：

```bash
cp conf/config.example.yaml conf/config.yaml
```

编辑 `conf/config.yaml`，配置七牛云参数：

```yaml
qiniu:
  accessKey: "your_qiniu_access_key"     # 七牛云 AccessKey
  secretKey: "your_qiniu_secret_key"     # 七牛云 SecretKey
  bucket: "your_bucket_name"             # 存储空间名称
  domain: "your_domain.com"              # CDN 域名
  zone: "z0"                             # 存储区域
```

### 启动微服务

在不同终端中启动各个服务：

```bash
# 终端1: 启动 User 服务
go run app/user/cmd/main.go

# 终端2: 启动 Task 服务
go run app/task/cmd/main.go

# 终端3: 启动 Files 服务
go run app/files/cmd/main.go

# 终端4: 启动 Gateway 服务
go run app/gateway/cmd/main.go
```

### 验证部署

访问健康检查接口：

```bash
# 检查 Gateway 服务
curl http://localhost:4000/ping

# 检查服务注册状态
curl http://localhost:2379/v2/keys/services
```

## 🧪 测试指南

项目提供了完整的测试套件，位于 `test/` 目录：

### 功能测试

```bash
cd test

# 七牛云表单上传测试
go run main.go  # 修改 main() 函数选择测试项

# 七牛云流式上传测试
go run stream_upload_test.go

# 全盘搜索和跨用户下载测试
go run global_search_test.go

# 文件删除测试
go run qiniu_delete_test.go
```

### Shell 脚本测试

```bash
# 下载功能测试
chmod +x simple_download_test.sh
./simple_download_test.sh

# 搜索功能测试
chmod +x global_search_test.sh
./global_search_test.sh

# 删除功能测试
chmod +x qiniu_delete_test.sh
./qiniu_delete_test.sh
```

### Postman 测试

导入 `test/Qiniu_Download_Test.postman_collection.json` 到 Postman 进行接口测试。

## 📊 性能特性

### 文件上传性能

| 上传方式   | 适用场景      | 文件大小限制 | 并发支持 | 特殊功能           |
| ---------- | ------------- | ----------- | -------- | ------------------ |
| 表单上传   | 小文件        | < 10MB      | 高       | 秒传、批量上传     |
| 流式上传   | 大文件        | 无限制       | 中       | 分片、断点续传     |
| 异步上传   | 批量处理      | < 10MB     | 极高     | 队列缓冲、重试机制 |

### 秒传机制

- **哈希算法**: SHA256
- **检测范围**: 跨用户全局检测
- **存储优化**: 相同文件只存储一份物理文件
- **用户体验**: 每个用户都有独立的文件记录

### 系统容量

- **并发用户**: 1000+ (基于 Gin + gRPC)
- **文件存储**: 无限制 (七牛云对象存储)
- **数据库**: 支持分库分表扩展
- **消息队列**: Kafka 支持水平扩展

## 🔧 开发指南

### 项目结构

```
grpc-todolist-disk/
├── app/                    # 微服务应用
│   ├── gateway/           # API 网关服务
│   │   ├── cmd/          # 启动入口
│   │   ├── http/         # HTTP 处理器
│   │   ├── rpc/          # RPC 客户端
│   │   └── router/       # 路由配置
│   ├── user/             # 用户服务
│   ├── task/             # 任务服务
│   └── files/            # 文件服务
│       ├── cmd/          # 启动入口
│       ├── dao/          # 数据访问层
│       ├── internal/     # 内部逻辑
│       └── utils/        # 工具函数
├── conf/                  # 配置文件
├── idl/                   # Protocol Buffers 定义
├── utils/                 # 公共工具
│   ├── ctl/              # 控制器工具
│   ├── e/                # 错误码定义
│   └── qiniu/            # 七牛云 SDK 封装
├── test/                  # 测试文件
├── docker-compose.yml     # Docker 编排
└── Makefile              # 构建脚本
```

### 添加新功能

1. **定义 Proto 接口**
   ```bash
   # 编辑 idl/files.proto
   # 重新生成代码
   make
   ```

2. **实现服务层逻辑**
   ```go
   // app/files/internal/service/files.go
   func (*FilesSrv) NewFunction(ctx context.Context, req *pb.Request) (*pb.Response, error) {
       // 实现业务逻辑
   }
   ```

3. **添加 HTTP 接口**
   ```go
   // app/gateway/http/files.go
   func NewHandler(ctx *gin.Context) {
       // HTTP 处理逻辑
   }
   ```

4. **配置路由**
   ```go
   // app/gateway/router/router.go
   authed.POST("new_endpoint", http.NewHandler)
   ```

### 代码规范

- **命名规范**: 遵循 Go 官方命名规范
- **错误处理**: 统一使用 `utils/e` 包定义的错误码
- **日志记录**: 使用结构化日志，包含请求 ID
- **配置管理**: 所有配置项都应在 `conf/config.yaml` 中定义

## 🐛 故障排除

### 常见问题

#### 1. 服务启动失败
```bash
# 检查端口占用
netstat -tulpn | grep :4000

# 检查 etcd 连接
curl http://localhost:2379/health
```

#### 2. 七牛云上传失败
- 检查 AccessKey 和 SecretKey 是否正确
- 确认存储空间名称和区域配置
- 验证域名是否已备案

#### 3. 数据库连接错误
```bash
# 检查 MySQL 服务状态
docker-compose ps mysql

# 查看数据库日志
docker-compose logs mysql
```

#### 4. Kafka 消息丢失
```bash
# 检查 Kafka 服务状态
docker-compose ps kafka

# 查看消费者组状态
docker exec -it kafka kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list
```

### 日志分析

```bash
# 查看服务日志
tail -f logs/gateway.log
tail -f logs/files.log

# 搜索错误日志
grep "ERROR" logs/*.log

# 分析请求链路
grep "request_id=xxx" logs/*.log
```

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. **Fork 项目**
2. **创建功能分支** (`git checkout -b feature/AmazingFeature`)
3. **提交更改** (`git commit -m 'Add some AmazingFeature'`)
4. **推送分支** (`git push origin feature/AmazingFeature`)
5. **创建 Pull Request**

### 提交规范

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 代码重构
test: 测试相关
chore: 构建过程或辅助工具的变动
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [gRPC-Go](https://github.com/grpc/grpc-go) - gRPC Go 实现
- [GORM](https://github.com/go-gorm/gorm) - Go ORM 库
- [七牛云](https://www.qiniu.com/) - 对象存储服务
- [Kafka](https://kafka.apache.org/) - 分布式消息队列

⭐ 如果这个项目对你有帮助，请给个 Star 支持一下！
