# Claude Relay Service MVP

Claude Code OAuth认证 + API请求转发的最小可行产品实现。

## 🎯 核心功能

1. **Claude Code OAuth认证** - 完整的OAuth 2.0 + PKCE流程
2. **智能请求转发** - 支持代理配置的Claude API请求转发
3. **Token自动管理** - 自动刷新过期的访问令牌
4. **代理支持** - 支持SOCKS5和HTTP/HTTPS代理

## 🚀 快速开始

### 1. 编译和运行

```bash
# 设置Go环境（如果需要）
export GOROOT=/Users/alfons/go/go1.24.4

# 编译
go build -o claude-relay ./cmd/server

# 运行
./claude-relay
```

### 2. 完成OAuth认证

访问 `http://localhost:3000` 查看API说明，然后按以下步骤操作：

#### 步骤1: 生成授权URL
```bash
curl -X POST http://localhost:3000/oauth/auth-url \
  -H "Content-Type: application/json" \
  -d '{}'
```

#### 步骤2: 完成授权
1. 复制返回的 `auth_url` 到浏览器
2. 登录Claude Code账号并授权
3. 从重定向URL中提取 `authorization_code`

#### 步骤3: 交换Token
```bash
curl -X POST http://localhost:3000/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "authorization_code": "你的授权码",
    "state": "返回的state值",
    "account_name": "my_account"
  }'
```

### 3. 测试API转发

```bash
curl -X POST "http://localhost:3000/api/v1/messages?account=my_account" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 100,
    "messages": [
      {
        "role": "user",
        "content": "Hello, Claude!"
      }
    ]
  }'
```

## 🧪 自动化测试

运行包含的测试脚本：

```bash
go run test/test_oauth.go
```

测试脚本将引导您完成完整的OAuth流程和API测试。

## 📁 项目结构

```
go-core/
├── cmd/server/          # 服务器主入口
├── internal/
│   ├── config/         # 配置管理
│   ├── oauth/          # OAuth认证模块
│   └── proxy/          # 代理和转发模块
├── test/               # 测试脚本
└── data/               # OAuth数据存储目录（运行时创建）
```

## 🔧 配置选项

可通过环境变量配置：

```bash
export PORT=3000                    # 服务端口
export HOST=0.0.0.0                # 服务主机
export CLAUDE_TIMEOUT=30s          # Claude API超时
export PROXY_TIMEOUT=30s           # 代理超时

# 全局代理配置（用于所有Claude Code服务器请求）
export GLOBAL_PROXY_ENABLED=true   # 启用全局代理
export GLOBAL_PROXY_TYPE=socks5    # 代理类型: socks5, http, https
export GLOBAL_PROXY_HOST=127.0.0.1 # 代理服务器地址
export GLOBAL_PROXY_PORT=1080      # 代理端口
export GLOBAL_PROXY_USERNAME=user  # 代理用户名（可选）
export GLOBAL_PROXY_PASSWORD=pass  # 代理密码（可选）
```

参考 `.env.example` 文件查看完整的配置示例。

## 🌐 代理配置

### 全局代理配置（推荐）

通过环境变量配置全局代理，所有到Claude Code服务器的请求都会自动使用此代理：

```bash
# 启用全局代理
export GLOBAL_PROXY_ENABLED=true
export GLOBAL_PROXY_TYPE=socks5        # 支持: socks5, http, https
export GLOBAL_PROXY_HOST=127.0.0.1
export GLOBAL_PROXY_PORT=1080
export GLOBAL_PROXY_USERNAME=username  # 可选
export GLOBAL_PROXY_PASSWORD=password  # 可选
```

### 请求特定代理配置

如果没有配置全局代理，可以在OAuth和API请求中单独指定代理：

```json
{
  "proxy_config": {
    "type": "socks5",           // 或 "http", "https"
    "host": "127.0.0.1",
    "port": 1080,
    "username": "user",         // 可选
    "password": "pass"          // 可选
  }
}
```

### 代理优先级

1. **全局代理** - 如果配置了全局代理，所有请求都会使用它
2. **请求特定代理** - 如果没有全局代理，使用请求中指定的代理
3. **直接连接** - 如果都没有配置，直接连接到服务器

## 📊 API端点

### OAuth管理
- `POST /oauth/auth-url` - 生成授权URL
- `POST /oauth/token` - 交换授权码获取token
- `GET /oauth/accounts` - 列出已认证账户
- `GET /oauth/accounts/:name/status` - 检查账户状态

### API转发
- `POST /api/v1/messages` - Claude消息API转发
- `GET /api/v1/models` - 模型列表（兼容性）

### 系统
- `GET /health` - 健康检查
- `GET /` - 服务信息和使用说明

## 🔍 故障排除

### 常见问题

1. **OAuth认证失败**
   - 确保授权码完整且未过期
   - 检查网络连接和代理设置

2. **Token刷新失败**
   - 检查refresh_token是否有效
   - 验证代理配置

3. **API转发失败**
   - 确认账户名正确
   - 检查token是否有效
   - 验证请求格式

### 调试信息

服务启动时会显示：
```
🚀 Claude Relay Service 启动成功
🌐 服务地址: http://0.0.0.0:3000
🔗 代理端点: http://0.0.0.0:3000/api/v1/messages
⚙️  OAuth管理: http://0.0.0.0:3000/oauth
```

### 数据存储

- OAuth数据存储在 `./data/oauth_账户名.json`
- PKCE临时数据存储在 `./data/pkce_state值.json`
- 所有敏感数据都是JSON格式，便于调试

## 🎯 MVP特性

这是一个最小可行产品，专注于核心功能：

✅ **已实现**
- 完整OAuth 2.0 + PKCE流程
- Token自动刷新机制
- 代理支持（SOCKS5/HTTP/HTTPS）
- 全局代理配置（优先级管理）
- 模块化API架构（handlers分离）
- 基本的API请求转发
- 简单的文件存储

🚧 **未包含**
- 数据库存储
- Web管理界面
- API Key认证
- 使用统计
- 多账户负载均衡
- 流式响应处理

这个MVP版本可以作为完整系统的基础，验证核心OAuth和转发功能。