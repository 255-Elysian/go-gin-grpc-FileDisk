# 七牛云上传测试

## 使用说明

### 1. 配置准备

在运行测试之前，请确保：

1. **七牛云配置**：在 `conf/config.yaml` 中正确配置七牛云参数
2. **服务启动**：确保以下服务正在运行
   - Files 服务 (端口: 10004)
   - Gateway 服务 (端口: 4000)
3. **用户认证**：更新测试代码中的 JWT token

### 2. 测试文件

当前测试文件路径：`C:\Users\elisia\Desktop\杂物\Linux 安装及使用.md`

如需修改，请在 `test/main.go` 中更新 `filePath` 变量。

### 3. 运行测试

```bash
cd test
go run main.go
```

### 4. 测试选项

在 `main()` 函数中可以选择不同的测试：

```go
func main() {
    //g()                  // 本地流式上传
    //k()                  // 异步上传
    //qiniuFileUpload()    // 七牛云表单上传
    qiniuBigFileUpload()   // 七牛云流式上传 (当前启用)
}
```

### 5. 测试功能

#### 七牛云表单上传 (`qiniuFileUpload`)
- 适用于小文件
- 使用 HTTP 表单上传
- 支持秒传功能

#### 七牛云流式上传 (`qiniuBigFileUpload`)
- 适用于大文件
- 使用 gRPC 流式上传
- 1MB 分片大小
- 支持秒传功能

### 6. 预期输出

#### 流式上传成功示例：
```
开始上传文件: C:\Users\elisia\Desktop\杂物\Linux 安装及使用.md
文件大小: 12345 bytes
发送分片 1: 12345 bytes, isLast: true

=== 上传完成 ===
响应码: 200
消息: 上传成功
文件ID: 123
访问URL: http://your-domain.com/uploads/5/1640995200000.md
总共发送: 1 个分片
总大小: 12345 bytes
```

#### 秒传成功示例：
```
=== 上传完成 ===
响应码: 200
消息: 秒传成功，文件已存在
文件ID: 124
访问URL: http://your-domain.com/uploads/5/1640995200000.md
总共发送: 1 个分片
总大小: 12345 bytes
```

### 7. 注意事项

1. **Token 更新**：测试代码中的 JWT token 可能已过期，请使用有效的 token
2. **文件路径**：确保测试文件存在且可读
3. **网络连接**：确保能连接到 gRPC 服务 (localhost:10004)
4. **七牛云配置**：确保七牛云配置正确，否则上传会失败

### 8. 故障排除

- **连接失败**：检查 Files 服务是否启动
- **认证失败**：更新 JWT token
- **上传失败**：检查七牛云配置和网络连接
- **文件不存在**：检查文件路径是否正确
