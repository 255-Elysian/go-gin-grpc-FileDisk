# 部署指南

## 环境要求

### 系统要求
- **操作系统**: Linux (推荐 Ubuntu 20.04+) / macOS / Windows
- **内存**: 最低 4GB，推荐 8GB+
- **磁盘**: 最低 20GB 可用空间
- **网络**: 稳定的互联网连接

### 软件依赖
- **Go**: 1.19+ 
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **Git**: 2.0+

## 开发环境部署

### 1. 克隆项目

```bash
git clone https://github.com/your-username/grpc-todolist-disk.git
cd grpc-todolist-disk
```

### 2. 启动基础设施

```bash
# 启动所有依赖服务
docker-compose up -d

# 验证服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f
```

启动的服务：
- **MySQL** (3306) - 数据库
- **Redis** (6379) - 缓存
- **etcd** (2379) - 服务注册
- **Kafka** (9092) - 消息队列
- **Zookeeper** (2181) - Kafka 依赖

### 3. 配置文件

```bash
# 复制配置模板
cp conf/config.example.yaml conf/config.yaml

# 编辑配置文件
vim conf/config.yaml
```

重要配置项：
```yaml
# 数据库配置
mysql:
  host: "localhost"
  port: "3306"
  database: "grpc-todolist-disk"
  username: "root"
  password: "123456"

# 七牛云配置
qiniu:
  accessKey: "your_access_key"
  secretKey: "your_secret_key"
  bucket: "your_bucket"
  domain: "your_domain.com"
  zone: "z0"
```

### 4. 数据库初始化

```bash
# 方式1: 自动创建表结构 (GORM AutoMigrate)
go run app/user/cmd/main.go    # 启动时会自动创建表

# 方式2: 手动执行 SQL
mysql -h localhost -u root -p123456 grpc-todolist-disk < scripts/init.sql
```

### 5. 启动微服务

```bash
# 方式1: 分别启动 (开发调试)
# 终端1
go run app/user/cmd/main.go

# 终端2  
go run app/task/cmd/main.go

# 终端3
go run app/files/cmd/main.go

# 终端4
go run app/gateway/cmd/main.go

# 方式2: 后台启动
nohup go run app/user/cmd/main.go > logs/user.log 2>&1 &
nohup go run app/task/cmd/main.go > logs/task.log 2>&1 &
nohup go run app/files/cmd/main.go > logs/files.log 2>&1 &
nohup go run app/gateway/cmd/main.go > logs/gateway.log 2>&1 &
```

### 6. 验证部署

```bash
# 检查服务健康状态
curl http://localhost:4000/ping

# 检查服务注册
curl http://localhost:2379/v2/keys/services

# 测试用户注册
curl -X POST http://localhost:4000/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

## 生产环境部署

### 1. 服务器准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装必要软件
sudo apt install -y git curl wget vim

# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.12.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 安装 Go
wget https://go.dev/dl/go1.20.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.20.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. 项目部署

```bash
# 创建项目目录
sudo mkdir -p /opt/grpc-todolist-disk
sudo chown $USER:$USER /opt/grpc-todolist-disk
cd /opt/grpc-todolist-disk

# 克隆项目
git clone https://github.com/your-username/grpc-todolist-disk.git .

# 构建项目
make build
```

### 3. 生产配置

```bash
# 生产配置文件
cp conf/config.example.yaml conf/config.prod.yaml

# 编辑生产配置
vim conf/config.prod.yaml
```

生产配置要点：
```yaml
server:
  port: ":4000"
  mode: "release"  # 生产模式

mysql:
  host: "your-mysql-host"
  port: "3306"
  database: "grpc_todolist_prod"
  username: "prod_user"
  password: "strong_password"

redis:
  address: "your-redis-host:6379"
  password: "redis_password"

qiniu:
  accessKey: "prod_access_key"
  secretKey: "prod_secret_key"
  bucket: "prod-bucket"
  domain: "cdn.yourdomain.com"
```

### 4. 系统服务配置

创建 systemd 服务文件：

```bash
# Gateway 服务
sudo tee /etc/systemd/system/grpc-gateway.service > /dev/null <<EOF
[Unit]
Description=gRPC Gateway Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/grpc-todolist-disk
ExecStart=/opt/grpc-todolist-disk/bin/gateway
Restart=always
RestartSec=5
Environment=CONFIG_PATH=/opt/grpc-todolist-disk/conf/config.prod.yaml

[Install]
WantedBy=multi-user.target
EOF

# 类似地创建其他服务...
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable grpc-gateway
sudo systemctl start grpc-gateway
sudo systemctl status grpc-gateway
```

### 5. Nginx 反向代理

```bash
# 安装 Nginx
sudo apt install -y nginx

# 配置文件
sudo tee /etc/nginx/sites-available/grpc-todolist > /dev/null <<EOF
server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://localhost:4000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # 文件上传大小限制
    client_max_body_size 100M;
}
EOF

# 启用站点
sudo ln -s /etc/nginx/sites-available/grpc-todolist /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 6. SSL 证书 (Let's Encrypt)

```bash
# 安装 Certbot
sudo apt install -y certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d yourdomain.com

# 自动续期
sudo crontab -e
# 添加: 0 12 * * * /usr/bin/certbot renew --quiet
```

## Docker 容器化部署

### 1. 构建镜像

```dockerfile
# Dockerfile.gateway
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway app/gateway/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gateway .
COPY --from=builder /app/conf ./conf
CMD ["./gateway"]
```

构建镜像：
```bash
docker build -f Dockerfile.gateway -t grpc-gateway:latest .
docker build -f Dockerfile.user -t grpc-user:latest .
docker build -f Dockerfile.task -t grpc-task:latest .
docker build -f Dockerfile.files -t grpc-files:latest .
```

### 2. Docker Compose 生产配置

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  gateway:
    image: grpc-gateway:latest
    ports:
      - "4000:4000"
    environment:
      - CONFIG_PATH=/app/conf/config.prod.yaml
    depends_on:
      - mysql
      - redis
      - etcd
    restart: unless-stopped

  user:
    image: grpc-user:latest
    ports:
      - "10002:10002"
    environment:
      - CONFIG_PATH=/app/conf/config.prod.yaml
    depends_on:
      - mysql
      - redis
      - etcd
    restart: unless-stopped

  # ... 其他服务配置
```

启动：
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 监控和日志

### 1. 日志管理

```bash
# 创建日志目录
mkdir -p /var/log/grpc-todolist

# 日志轮转配置
sudo tee /etc/logrotate.d/grpc-todolist > /dev/null <<EOF
/var/log/grpc-todolist/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 www-data www-data
}
EOF
```

### 2. 健康检查

```bash
# 健康检查脚本
#!/bin/bash
# health_check.sh

services=("gateway:4000" "user:10002" "task:10003" "files:10004")

for service in "${services[@]}"; do
    name=${service%:*}
    port=${service#*:}
    
    if curl -f -s http://localhost:$port/health > /dev/null; then
        echo "✅ $name service is healthy"
    else
        echo "❌ $name service is down"
        # 重启服务
        sudo systemctl restart grpc-$name
    fi
done
```

### 3. 性能监控

使用 Prometheus + Grafana：

```yaml
# monitoring/docker-compose.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

## 故障排除

### 常见问题

1. **端口冲突**
```bash
# 检查端口占用
netstat -tulpn | grep :4000
# 杀死进程
sudo kill -9 <PID>
```

2. **服务无法启动**
```bash
# 查看服务日志
sudo journalctl -u grpc-gateway -f
# 检查配置文件
go run app/gateway/cmd/main.go --config-check
```

3. **数据库连接失败**
```bash
# 测试数据库连接
mysql -h localhost -u root -p
# 检查防火墙
sudo ufw status
```

4. **七牛云上传失败**
```bash
# 测试网络连接
curl -I https://upload.qiniup.com
# 验证配置
go run scripts/test_qiniu.go
```

### 性能优化

1. **数据库优化**
```sql
-- 添加索引
CREATE INDEX idx_files_user_id ON files(user_id);
CREATE INDEX idx_files_created_at ON files(created_at);

-- 查询优化
EXPLAIN SELECT * FROM files WHERE user_id = 1;
```

2. **Redis 缓存**
```bash
# 监控 Redis 性能
redis-cli info stats
redis-cli monitor
```

3. **系统资源**
```bash
# 监控系统资源
htop
iotop
nethogs
```

## 备份和恢复

### 数据库备份

```bash
# 自动备份脚本
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
mysqldump -h localhost -u root -p123456 grpc-todolist-disk > backup_$DATE.sql
gzip backup_$DATE.sql

# 保留最近30天的备份
find /backup -name "backup_*.sql.gz" -mtime +30 -delete
```

### 配置备份

```bash
# 备份配置文件
tar -czf config_backup_$(date +%Y%m%d).tar.gz conf/
```

---

部署完成后，记得定期检查服务状态和性能指标！
