# Claude Relay Service Go 1.23 迁移详细Todo清单

## 阶段一：项目基础搭建

### 1.1 Go项目初始化
- [ ] **初始化Go模块**
  - 执行`go mod init claude-relay-go`
  - 设置Go版本为1.23
  - 配置go.sum文件

- [ ] **创建标准Go项目目录结构**
  ```
  cmd/
    server/
      main.go           # Web服务入口
  internal/
    api/
      handlers/         # HTTP处理器
      middleware/       # 中间件
      routes/          # 路由定义
    services/          # 业务逻辑服务
    models/           # 数据模型
    repository/       # 数据访问层
    utils/           # 工具函数
  pkg/               # 可导出的包
  configs/          # 配置文件
  scripts/         # 部署脚本
  ```

- [ ] **依赖管理配置**
  - 添加Gin框架：`github.com/gin-gonic/gin`
  - 添加GORM ORM：`gorm.io/gorm`
  - 添加SQLite驱动：`gorm.io/driver/sqlite`
  - 添加纯Go SQLite：`modernc.org/sqlite`
  - 添加配置管理：`github.com/spf13/viper`
  - 添加日志库：`go.uber.org/zap`
  - 添加JWT库：`github.com/golang-jwt/jwt/v5`
  - 添加限流库：`golang.org/x/time/rate`
  - 添加代理支持：`golang.org/x/net/proxy`
  - 添加嵌入式静态文件：Go标准库`embed`
  - 添加加密库：标准库`crypto/aes`
  - 添加测试库：`github.com/stretchr/testify`
  - 添加数据库迁移：`gorm.io/gorm/migration`

### 1.2 配置系统实现
- [ ] **配置结构体定义**
  - 定义`Config`结构体包含所有配置项
  - Server配置（port, host, nodeEnv, trustProxy）
  - Security配置（jwtSecret, adminSessionTimeout, apiKeyPrefix, encryptionKey）
  - SQLite配置（数据库文件路径、连接池设置等）
  - Claude API配置（apiUrl, apiVersion, betaHeader）
  - 代理配置（timeout, maxRetries）
  - 日志配置（level, dirname, maxSize, maxFiles）
  - Web界面配置（title, description, logoUrl等）

- [ ] **使用Viper读取配置**
  - 支持YAML、JSON、ENV格式
  - 环境变量覆盖机制
  - 配置文件监听和热重载
  - 配置验证和默认值设置

- [ ] **配置验证逻辑**
  - 必填字段验证
  - 格式验证（端口范围、URL格式等）
  - 安全性检查（密钥长度等）

### 1.3 日志系统实现
- [ ] **Zap日志配置**
  - 配置不同级别日志（DEBUG, INFO, WARN, ERROR）
  - 结构化日志格式（JSON格式用于生产环境）
  - 控制台彩色输出（开发环境）
  - 时间戳格式配置

- [ ] **文件轮转实现**
  - 按日期轮转日志文件
  - 文件大小限制和清理策略
  - 多个日志文件管理（error.log, access.log等）

- [ ] **日志元数据支持**
  - 请求ID追踪
  - 性能计时器实现
  - 上下文信息记录
  - 错误堆栈跟踪

- [ ] **日志统计和健康检查**
  - 日志写入统计
  - 日志系统健康状态检查
  - 日志级别动态调整

## 阶段二：核心服务迁移

### 2.1 SQLite数据库层实现
- [ ] **SQLite数据库连接管理**
  - 使用GORM ORM和SQLite驱动
  - 数据库文件路径配置
  - 连接池配置（空闲连接、最大连接数）
  - WAL模式配置提升并发性能
  - 数据库健康检查

- [ ] **GORM模型定义**
  - 定义所有数据模型结构体
  - ApiKey模型（ID, Name, Hash, Limits等）
  - ClaudeAccount模型（OAuth数据、代理配置）
  - AdminSession模型（会话管理）
  - AdminUser模型（管理员信息）
  - UsageRecord模型（使用统计）
  - SessionMapping模型（会话映射）
  - SystemConfig模型（系统配置）
  - RateLimitRecord模型（限流记录）

- [ ] **数据库迁移系统**
  - 自动迁移配置（AutoMigrate）
  - 数据库版本管理
  - 向前和向后兼容性
  - 初始数据种子文件

- [ ] **Repository接口定义**
  - 定义`ApiKeyRepository`接口
  - 定义`ClaudeAccountRepository`接口
  - 定义`AdminSessionRepository`接口
  - 定义`UsageRecordRepository`接口
  - 定义`SessionMappingRepository`接口

- [ ] **API Key相关Repository实现**
  - `CreateApiKey(apiKey *ApiKey)` - 创建API Key
  - `GetApiKeyByID(id string)` - 通过ID获取
  - `GetApiKeyByHash(hash string)` - 通过哈希快速查找
  - `UpdateApiKey(apiKey *ApiKey)` - 更新API Key
  - `DeleteApiKey(id string)` - 删除API Key
  - `ListApiKeys(filter *ApiKeyFilter)` - 获取API Key列表
  - `UpdateConcurrency(id string, delta int)` - 更新并发数

- [ ] **Claude账户Repository实现**
  - `CreateClaudeAccount(account *ClaudeAccount)` - 创建账户
  - `GetClaudeAccountByID(id string)` - 获取账户详情
  - `UpdateClaudeAccount(account *ClaudeAccount)` - 更新账户
  - `DeleteClaudeAccount(id string)` - 删除账户
  - `ListClaudeAccounts(filter *AccountFilter)` - 获取账户列表
  - `SetRateLimit(id string, until time.Time)` - 设置限流
  - `RemoveRateLimit(id string)` - 移除限流
  - `GetAvailableAccounts()` - 获取可用账户

- [ ] **会话管理Repository实现**
  - `CreateSession(session *AdminSession)` - 创建会话
  - `GetSessionByID(sessionId string)` - 获取会话
  - `UpdateSession(session *AdminSession)` - 更新会话
  - `DeleteSession(sessionId string)` - 删除会话
  - `CleanupExpiredSessions()` - 清理过期会话

- [ ] **使用统计Repository实现**
  - `RecordUsage(record *UsageRecord)` - 记录使用数据
  - `GetUsageStats(filter *UsageFilter)` - 获取使用统计
  - `GetDailyStats(date time.Time)` - 获取每日统计
  - `GetTokenUsage(keyId string, period string)` - 获取token使用量
  - `GetSystemStats()` - 获取系统统计

- [ ] **会话映射Repository实现**
  - `CreateSessionMapping(mapping *SessionMapping)` - 创建映射
  - `GetSessionMapping(hash string)` - 获取映射
  - `UpdateSessionMapping(mapping *SessionMapping)` - 更新映射
  - `DeleteSessionMapping(hash string)` - 删除映射
  - `CleanupExpiredMappings()` - 清理过期映射

### 2.2 加密工具实现
- [ ] **AES-256-CBC加密**
  - `Encrypt(data, key)` - 数据加密方法
  - `Decrypt(encryptedData, key)` - 数据解密方法
  - IV随机生成和管理
  - 密钥派生函数（PBKDF2）

- [ ] **哈希工具**
  - `SHA256Hash(data)` - SHA256哈希计算
  - `GenerateApiKey(prefix)` - API Key生成
  - `HashApiKey(apiKey)` - API Key哈希计算
  - 加盐哈希实现

### 2.3 认证系统实现
- [ ] **API Key验证中间件**
  - 从请求头提取API Key（x-api-key, Authorization, api-key）
  - API Key格式验证（长度、前缀检查）
  - 通过SQLite索引快速验证（使用api_key_hash唯一索引）
  - API Key状态检查（活跃、过期、禁用）
  - 使用配额检查和限流
  - 内存缓存热点API Key减少数据库查询

- [ ] **并发控制中间件**
  - 数据库事务安全的并发数增减
  - 请求结束时自动清理并发计数
  - 支持多种请求结束事件监听
  - 并发超限的错误响应
  - 内存计数器 + 定期数据库同步

- [ ] **速率限制中间件**
  - 基于Token Bucket算法的限流
  - 支持按API Key的个性化限流配置
  - SQLite + 内存混合存储限流状态
  - 限流信息在响应头中返回
  - 滑动窗口限流算法实现

- [ ] **管理员认证中间件**
  - JWT token验证
  - 会话有效性检查
  - 会话自动续期机制
  - 安全上下文设置

### 2.4 Claude账户管理服务
- [ ] **OAuth PKCE实现**
  - `GenerateState()` - 随机state生成
  - `GenerateCodeVerifier()` - PKCE code verifier生成
  - `GenerateCodeChallenge(verifier)` - code challenge计算
  - `GenerateAuthorizeURL(challenge, state, proxy)` - 授权URL生成
  - 代理支持的HTTP客户端配置

- [ ] **Token管理**
  - `ExchangeCodeForToken(code, verifier, proxy)` - 授权码换取token
  - `RefreshAccessToken(refreshToken, proxy)` - 刷新访问token
  - `IsTokenExpired(token)` - token过期检查
  - `GetValidAccessToken(accountId)` - 获取有效访问token
  - 自动刷新机制（提前10秒）

- [ ] **账户选择算法**
  - `SelectAccountForApiKey(apiKeyData, sessionHash)` - 账户选择
  - 专属绑定账户支持
  - Sticky会话实现
  - 负载均衡算法
  - 限流账户自动排除

- [ ] **账户状态管理**
  - `MarkAccountRateLimited(accountId, sessionHash)` - 标记限流
  - `RemoveAccountRateLimit(accountId)` - 移除限流
  - `IsAccountRateLimited(accountId)` - 检查限流状态
  - `CleanupErrorAccounts()` - 清理错误账户

## 阶段三：API服务迁移

### 3.1 HTTP路由框架搭建
- [ ] **Gin路由器配置**
  - 创建Gin引擎实例
  - 中间件栈配置顺序
  - 路由组织和分组（/api/v1, /admin, /web）
  - 错误处理中间件

- [ ] **安全中间件配置**
  - CORS中间件配置
  - 安全头设置（Content-Security-Policy等）
  - 请求大小限制
  - 请求超时配置

- [ ] **请求日志中间件**
  - HTTP请求日志记录
  - 请求ID生成和追踪
  - 性能计时（请求处理时间）
  - 错误请求特殊标记

### 3.2 客户端API路由实现
- [ ] **消息处理端点（/api/v1/messages）**
  - POST方法处理器实现
  - 请求体解析和验证
  - Claude API请求转发
  - 流式响应支持（SSE格式）
  - 使用统计提取和记录
  - 错误处理和格式化

- [ ] **模型列表端点（/api/v1/models）**
  - GET方法处理器
  - 兼容OpenAI API格式的模型列表
  - 静态模型信息返回

- [ ] **使用统计端点（/api/v1/usage）**
  - GET方法处理器
  - 时间范围参数解析
  - 按API Key过滤统计
  - 多维度统计数据聚合

- [ ] **API Key信息端点（/api/v1/key-info）**
  - GET方法处理器
  - API Key基本信息返回
  - 使用统计摘要
  - 配额和限制信息

### 3.3 代理转发服务核心实现
- [ ] **Claude API请求转发**
  - HTTP客户端配置（超时、重试）
  - 代理配置支持（SOCKS5/HTTP）
  - 请求头处理和转发
  - 请求体处理和验证

- [ ] **流式响应处理**
  - SSE（Server-Sent Events）实现
  - 数据流实时转发
  - 客户端断开检测
  - 资源清理机制
  - 使用量从流中提取

- [ ] **会话管理**
  - 会话哈希生成算法
  - Sticky会话映射管理
  - 会话状态跟踪
  - 会话清理机制

- [ ] **错误处理和重试**
  - 限流错误检测和处理
  - 自动重试机制（指数退避）
  - 错误响应格式化
  - 失败账户标记

### 3.4 管理后台API实现
- [ ] **Claude账户管理接口**
  - `POST /admin/claude-accounts` - 创建账户
  - `GET /admin/claude-accounts` - 获取账户列表
  - `GET /admin/claude-accounts/:id` - 获取账户详情
  - `PUT /admin/claude-accounts/:id` - 更新账户
  - `DELETE /admin/claude-accounts/:id` - 删除账户
  - `POST /admin/claude-accounts/test/:id` - 测试账户

- [ ] **OAuth集成接口**
  - `POST /admin/claude-accounts/generate-auth-url` - 生成授权URL
  - `POST /admin/claude-accounts/exchange-code` - 交换授权码
  - `POST /admin/claude-accounts/refresh-token/:id` - 刷新token

- [ ] **API Key管理接口**
  - `POST /admin/api-keys` - 创建API Key
  - `GET /admin/api-keys` - 获取API Key列表
  - `GET /admin/api-keys/:id` - 获取API Key详情
  - `PUT /admin/api-keys/:id` - 更新API Key
  - `DELETE /admin/api-keys/:id` - 删除API Key
  - `POST /admin/api-keys/:id/reset` - 重置API Key

- [ ] **系统统计接口**
  - `GET /admin/dashboard` - 仪表板数据
  - `GET /admin/usage-stats` - 使用统计
  - `GET /admin/system-info` - 系统信息
  - `GET /admin/logs` - 系统日志

### 3.5 Web界面路由实现
- [ ] **静态文件服务**
  - 白名单文件机制实现
  - 文件存在性检查
  - MIME类型设置
  - 缓存控制头设置
  - 安全文件访问验证

- [ ] **管理员认证接口**
  - `POST /web/auth/login` - 管理员登录
  - `POST /web/auth/logout` - 管理员登出
  - `GET /web/auth/user` - 获取用户信息
  - `POST /web/auth/refresh` - 刷新token
  - `POST /web/auth/change-password` - 修改密码

- [ ] **会话管理**
  - 会话创建和存储
  - 会话验证和刷新
  - 会话清理和过期处理
  - 安全注销机制

## 阶段四：Web界面完善

### 4.1 静态资源服务
- [ ] **嵌入式静态文件**
  - 使用Go embed包嵌入前端资源
  - 编译时将web目录打包到二进制
  - 支持开发模式的外部文件服务
  - 生产模式的嵌入式文件服务

- [ ] **文件服务优化**
  - 允许文件清单定义
  - 文件路径验证和规范化
  - 目录遍历攻击防护
  - 文件类型验证
  - 条件请求支持（If-Modified-Since）
  - ETag生成和验证
  - 压缩支持（Gzip）
  - 缓存策略配置

### 4.2 Web管理功能增强
- [ ] **系统状态监控Web界面**
  - 实时系统状态显示
  - 数据库连接状态监控
  - API Key和Claude账户统计
  - 系统资源使用图表
  - 实时日志查看界面

- [ ] **数据迁移Web工具**
  - Redis到SQLite迁移界面
  - 迁移进度实时显示
  - 迁移验证结果展示
  - 错误处理和日志显示

- [ ] **管理界面优化**
  - 操作确认对话框
  - 批量操作支持
  - 数据导入导出功能
  - 系统配置界面
  - 用户操作日志

## 阶段五：部署和优化

### 5.1 健康检查和监控
- [ ] **健康检查端点**
  - `GET /health` - 基础健康检查
  - `GET /metrics` - Prometheus格式指标
  - 组件健康状态检查（Redis、日志系统）
  - 系统资源监控（内存、CPU、磁盘）

- [ ] **性能指标收集**
  - HTTP请求计数和延迟
  - Redis连接和操作统计
  - 内存使用和GC统计
  - Goroutine数量监控

- [ ] **告警和通知**
  - 关键错误告警机制
  - 性能阈值监控
  - 系统异常检测
  - 日志错误统计

### 5.2 Docker化部署
- [ ] **Dockerfile优化**
  - 多阶段构建配置
  - 基础镜像选择（Alpine Linux）
  - 构建层缓存优化
  - 安全用户配置

- [ ] **Docker Compose配置**
  - 服务编排配置
  - 环境变量管理
  - 卷挂载配置
  - 网络配置

- [ ] **容器健康检查**
  - 健康检查脚本
  - 容器重启策略
  - 资源限制配置
  - 日志收集配置

### 5.3 性能优化
- [ ] **并发处理优化**
  - Goroutine池管理
  - 连接池配置调优
  - 内存池使用
  - GC参数调优

- [ ] **缓存策略优化**
  - API Key验证结果缓存
  - 账户信息缓存
  - 静态文件缓存
  - Redis管道操作

- [ ] **资源管理优化**
  - 内存使用监控和优化
  - 文件描述符管理
  - 网络连接复用
  - 垃圾回收优化

## 阶段六：数据迁移工具开发

### 6.1 Redis到SQLite迁移工具
- [ ] **数据迁移Web接口**
  - Web管理界面中的迁移功能
  - 迁移进度WebSocket实时推送
  - 迁移结果验证和展示
  - 错误处理和日志显示

- [ ] **API Key数据迁移**
  - 从Redis `apikey:*` 键读取数据
  - 解析JSON数据结构
  - 验证数据完整性
  - 插入到SQLite `api_keys` 表
  - 重建哈希映射索引

- [ ] **Claude账户数据迁移**
  - 从Redis `claude_account:*` 键读取加密数据
  - 保持AES加密一致性
  - 迁移OAuth令牌和代理配置
  - 状态和错误信息迁移

- [ ] **会话数据迁移**
  - 管理员会话数据迁移
  - 会话映射关系迁移
  - 过期时间处理

- [ ] **使用统计数据迁移**
  - 从Redis `usage:*` 键提取历史数据
  - 按时间范围批量处理
  - 数据去重和验证
  - 统计数据重新计算

- [ ] **迁移验证和回滚**
  - 数据完整性检查
  - 记录数量对比
  - 关键数据采样验证
  - 失败回滚机制

## 阶段七：测试和验证

### 7.1 单元测试
- [ ] **服务层测试**
  - 每个服务方法独立测试
  - Mock SQLite数据库测试
  - 错误处理测试
  - 边界条件测试

- [ ] **工具函数测试**
  - 加密解密函数测试
  - 哈希函数测试
  - 配置解析测试
  - 日志功能测试

### 7.2 集成测试
- [ ] **API端到端测试**
  - 完整请求流程测试
  - 错误场景测试
  - 并发请求测试
  - 长时间运行测试

- [ ] **数据库集成测试**
  - SQLite数据库操作集成测试
  - GORM事务操作测试
  - 数据一致性测试
  - 数据库锁和并发测试
  - 迁移脚本测试

### 7.3 性能测试
- [ ] **压力测试**
  - 高并发请求测试
  - 内存泄漏检测
  - CPU使用率监控
  - 响应时间分析

- [ ] **基准测试**
  - 与Node.js版本性能对比
  - 吞吐量对比测试
  - 资源使用对比
  - 启动时间对比

### 7.4 兼容性测试
- [ ] **API兼容性**
  - 现有客户端兼容性测试
  - 响应格式验证
  - 错误消息一致性
  - HTTP状态码一致性

- [ ] **数据兼容性**
  - SQLite数据结构与Redis数据对应性验证
  - 加密数据解密验证
  - 配置文件兼容性
  - 会话数据兼容性
  - Redis到SQLite数据迁移工具测试

## 完成标准

### 功能完整性验证
- [ ] 所有现有API端点完全实现
- [ ] 所有管理功能正常工作
- [ ] Web管理界面功能完整
- [ ] Web界面正常访问

### 性能要求达标
- [ ] 内存使用降低30%以上
- [ ] 并发处理能力提升2倍以上
- [ ] 响应时间优化15%以上
- [ ] 启动时间优化50%以上

### 部署就绪检查
- [ ] Docker镜像构建成功
- [ ] 配置文件迁移完成
- [ ] 数据迁移脚本验证
- [ ] 监控和告警配置

### 文档和维护
- [ ] API文档更新
- [ ] 部署文档编写
- [ ] 故障排除指南
- [ ] 性能调优指南

此详细Todo清单确保了Go迁移项目的每个功能点都有明确的实现要求和验证标准。