# API 接口文档

## 基础信息

- **Base URL**: `http://localhost:4000`
- **认证方式**: JWT Bearer Token
- **Content-Type**: `application/json` (除文件上传外)

## 认证接口

### 用户注册

**接口**: `POST /api/v1/user/register`

**请求参数**:
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success",
    "user_id": 1
  },
  "msg": "success"
}
```

### 用户登录

**接口**: `POST /api/v1/user/login`

**请求参数**:
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success",
    "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": 1
  },
  "msg": "success"
}
```

## 文件接口

### 七牛云表单上传

**接口**: `POST /api/v1/qiniu_file_upload`

**请求头**:
```
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

**请求参数**:
- `file`: 文件数据 (form-data)

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success",
    "file_id": 123,
    "object_url": "http://domain.com/uploads/file.jpg"
  },
  "msg": "success"
}
```

### 七牛云流式上传

**接口**: `POST /api/v1/qiniu_big_file_upload`

**说明**: 使用 gRPC 流式接口，适用于大文件上传

### 全盘文件搜索

**接口**: `GET /api/v1/global_file_search`

**请求头**:
```
Authorization: Bearer <jwt_token>
```

**查询参数**:
- `file_name`: 文件名关键词 (可选)
- `page`: 页码，从1开始 (可选，默认1)
- `page_size`: 每页大小 (可选，默认10，最大100)
- `bucket`: 存储桶过滤 (可选，如"qiniu")

**请求示例**:
```
GET /api/v1/global_file_search?file_name=test&page=1&page_size=10&bucket=qiniu
```

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success",
    "files": [
      {
        "file_id": 123,
        "file_name": "test.jpg",
        "file_size": 1024000,
        "bucket": "qiniu",
        "object_name": "http://domain.com/uploads/test.jpg",
        "file_hash": "abc123...",
        "user_id": 5,
        "created_at": "2024-01-01 12:00:00",
        "updated_at": "2024-01-01 12:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  },
  "msg": "success"
}
```

### 文件下载

**接口**: `GET /api/v1/qiniu_file_download`

**请求头**:
```
Authorization: Bearer <jwt_token>
```

**查询参数**:
- `file_id`: 文件ID (必填)
- `user_id`: 用户ID (可选，不填支持跨用户下载)

**请求示例**:
```
GET /api/v1/qiniu_file_download?file_id=123
```

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success",
    "download_url": "http://domain.com/uploads/file.jpg",
    "file_name": "test.jpg"
  },
  "msg": "success"
}
```

### 文件删除

**接口**: `DELETE /api/v1/qiniu_file_delete`

**请求头**:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**请求参数**:
```json
{
  "file_id": 123
}
```

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success"
  },
  "msg": "success"
}
```

### 文件列表

**接口**: `GET /api/v1/file_list`

**请求头**:
```
Authorization: Bearer <jwt_token>
```

**查询参数**:
- `page`: 页码 (可选，默认1)
- `page_size`: 每页大小 (可选，默认10)

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success",
    "files": [
      {
        "file_id": 123,
        "file_name": "test.jpg",
        "file_size": 1024000,
        "created_at": "2024-01-01 12:00:00"
      }
    ],
    "total": 1
  },
  "msg": "success"
}
```

## 备忘录接口

### 创建备忘录

**接口**: `POST /api/v1/task`

**请求头**:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**请求参数**:
```json
{
  "title": "学习 Go 语言",
  "content": "完成 gRPC 项目开发",
  "status": 0
}
```

**状态说明**:
- `0`: 待办
- `1`: 进行中
- `2`: 已完成

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success",
    "task_id": 456
  },
  "msg": "success"
}
```

### 获取备忘录列表

**接口**: `GET /api/v1/task`

**请求头**:
```
Authorization: Bearer <jwt_token>
```

**查询参数**:
- `page`: 页码 (可选，默认1)
- `page_size`: 每页大小 (可选，默认10)

**响应示例**:
```json
{
  "status": 200,
  "data": {
    "code": 200,
    "msg": "success",
    "tasks": [
      {
        "task_id": 456,
        "title": "学习 Go 语言",
        "content": "完成 gRPC 项目开发",
        "status": 0,
        "created_at": "2024-01-01 12:00:00",
        "updated_at": "2024-01-01 12:00:00"
      }
    ],
    "total": 1
  },
  "msg": "success"
}
```

### 更新备忘录

**接口**: `PUT /api/v1/task`

**请求头**:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**请求参数**:
```json
{
  "task_id": 456,
  "title": "学习 Go 语言 (更新)",
  "content": "完成 gRPC 项目开发和部署",
  "status": 1
}
```

### 删除备忘录

**接口**: `DELETE /api/v1/task`

**请求头**:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**请求参数**:
```json
{
  "task_id": 456
}
```

## 错误码说明

| 错误码 | 说明           |
| ------ | -------------- |
| 200    | 成功           |
| 400    | 请求参数错误   |
| 401    | 未授权         |
| 403    | 权限不足       |
| 404    | 资源不存在     |
| 500    | 服务器内部错误 |

## 使用示例

### curl 示例

```bash
# 用户登录
curl -X POST http://localhost:4000/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 文件上传
curl -X POST http://localhost:4000/api/v1/qiniu_file_upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/file.jpg"

# 文件搜索
curl -X GET "http://localhost:4000/api/v1/global_file_search?file_name=test" \
  -H "Authorization: Bearer <token>"
```

### JavaScript 示例

```javascript
// 登录
const loginResponse = await fetch('/api/v1/user/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username: 'test', password: '123456' })
});

// 文件上传
const formData = new FormData();
formData.append('file', fileInput.files[0]);

const uploadResponse = await fetch('/api/v1/qiniu_file_upload', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: formData
});
```
