# 🚀 gRPC 分布式文件存储系统

一个基于 Go 的企业级分布式文件存储系统，集成备忘录管理和云存储网盘功能。

## ✨ 主要特性

- 🔐 **完整的用户系统** - 注册、登录、JWT 认证
- 📝 **备忘录管理** - 任务创建、分类、状态管理
- 📁 **多种上传方式** - 表单上传、流式上传、异步上传
- ⚡ **智能秒传** - 基于 SHA256 的跨用户秒传
- 🔍 **全盘搜索** - 文件名模糊搜索，跨用户文件发现
- ☁️ **七牛云集成** - 对象存储 + CDN 加速
- 🏗️ **微服务架构** - 4 个独立服务模块
- 📊 **异步处理** - Kafka 消息队列

## 🏗️ 系统架构

```
Gateway (4000) ←→ User (10002)
       ↓              ↓
    Files (10004) ←→ Task (10003)
       ↓
   七牛云存储
```

**基础设施**: MySQL + Redis + Kafka + etcd

## 🚀 快速开始

### 1. 环境准备
- Go 1.19+
- Docker & Docker Compose

### 2. 启动基础服务
```bash
git clone <repository>
cd grpc-todolist-disk
docker-compose up -d
```

### 3. 配置七牛云
```bash
cp conf/config.example.yaml conf/config.yaml
# 编辑配置文件，填入七牛云信息
```

### 4. 启动微服务
```bash
# 4个终端分别启动
go run app/user/cmd/main.go     # 用户服务
go run app/task/cmd/main.go     # 任务服务  
go run app/files/cmd/main.go    # 文件服务
go run app/gateway/cmd/main.go  # 网关服务
```

## 📖 主要接口

### 用户认证
```bash
# 注册
POST /api/v1/user/register
{"username": "test", "password": "123456"}

# 登录
POST /api/v1/user/login  
{"username": "test", "password": "123456"}
```

### 文件操作
```bash
# 上传文件
POST /api/v1/qiniu_file_upload
Content-Type: multipart/form-data

# 搜索文件
GET /api/v1/global_file_search?file_name=test

# 下载文件
GET /api/v1/qiniu_file_download?file_id=123

# 删除文件
DELETE /api/v1/qiniu_file_delete
{"file_id": 123}
```

### 备忘录
```bash
# 创建任务
POST /api/v1/task
{"title": "学习Go", "content": "完成项目", "status": 0}

# 获取任务列表
GET /api/v1/task?page=1&page_size=10
```

## 🧪 测试

```bash
cd test

# Go 测试程序
go run main.go                    # 基础功能测试
go run global_search_test.go      # 搜索测试
go run qiniu_delete_test.go       # 删除测试

# Shell 脚本测试
./simple_download_test.sh         # 下载测试
./global_search_test.sh           # 搜索测试
```

## 📁 项目结构

```
├── app/                 # 微服务
│   ├── gateway/        # API网关
│   ├── user/           # 用户服务
│   ├── task/           # 任务服务
│   └── files/          # 文件服务
├── conf/               # 配置文件
├── idl/                # Proto定义
├── utils/              # 公共工具
├── test/               # 测试文件
└── docker-compose.yml  # Docker编排
```

## 🔧 核心功能

### 文件上传
- **表单上传**: 小文件 < 100MB
- **流式上传**: 大文件分片上传
- **异步上传**: Kafka 队列处理
- **智能秒传**: SHA256 哈希检测

### 存储特性
- **跨用户共享**: 相同文件只存一份
- **全盘搜索**: 支持文件名模糊搜索
- **安全删除**: 智能判断是否删除物理文件
- **云存储**: 七牛云 CDN 加速访问

## 🐛 常见问题

### 服务启动失败
```bash
# 检查端口占用
netstat -tulpn | grep :4000

# 检查依赖服务
docker-compose ps
```

### 七牛云配置
- 确认 AccessKey/SecretKey 正确
- 检查存储空间和域名配置
- 验证域名备案状态

### 数据库问题
```bash
# 查看MySQL状态
docker-compose logs mysql

# 重启数据库
docker-compose restart mysql
```

## 📊 性能指标

- **并发用户**: 1000+
- **文件存储**: 无限制 (七牛云)
- **上传速度**: 取决于网络带宽
- **秒传响应**: < 100ms

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

---

⭐ 觉得有用请给个 Star！
