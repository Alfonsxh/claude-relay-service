# Claude Relay Service SQLite 数据模型设计

## 数据库架构概述

将现有Redis键值存储迁移到SQLite关系型数据库，保持数据完整性和功能一致性。

## 核心数据表设计

### 1. API Keys 表 (api_keys)

替代Redis的 `apikey:*` 键

```sql
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY,           -- UUID格式
    name TEXT NOT NULL,            -- API Key名称
    api_key_hash TEXT UNIQUE NOT NULL, -- SHA256哈希值
    is_active BOOLEAN DEFAULT TRUE,     -- 是否激活
    token_limit INTEGER DEFAULT 1000000, -- token限制
    concurrency_limit INTEGER DEFAULT 0, -- 并发限制
    current_concurrency INTEGER DEFAULT 0, -- 当前并发数
    dedicated_account_id TEXT,      -- 专属绑定账户ID
    description TEXT DEFAULT '',   -- 描述
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,           -- 过期时间
    last_used_at DATETIME,         -- 最后使用时间
    
    FOREIGN KEY (dedicated_account_id) REFERENCES claude_accounts(id)
);

-- 索引
CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(api_key_hash);
CREATE INDEX idx_api_keys_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_dedicated ON api_keys(dedicated_account_id);
```

### 2. Claude账户表 (claude_accounts)

替代Redis的 `claude_account:*` 键

```sql
CREATE TABLE claude_accounts (
    id TEXT PRIMARY KEY,           -- UUID格式
    name TEXT NOT NULL,            -- 账户名称
    email TEXT,                    -- 关联邮箱
    is_active BOOLEAN DEFAULT TRUE, -- 是否激活
    is_shared BOOLEAN DEFAULT TRUE, -- 是否共享账户
    
    -- OAuth数据（加密存储）
    encrypted_oauth_data TEXT,     -- 加密的OAuth数据（JSON）
    oauth_scopes TEXT,             -- OAuth权限范围
    oauth_expires_at DATETIME,     -- OAuth过期时间
    
    -- 代理配置
    proxy_type TEXT,               -- 代理类型：socks5/http/none
    proxy_host TEXT,               -- 代理主机
    proxy_port INTEGER,            -- 代理端口
    proxy_username TEXT,           -- 代理用户名
    proxy_password_encrypted TEXT, -- 加密的代理密码
    
    -- 状态管理
    is_rate_limited BOOLEAN DEFAULT FALSE, -- 是否限流
    rate_limit_until DATETIME,     -- 限流结束时间
    last_error TEXT,               -- 最后错误信息
    error_count INTEGER DEFAULT 0, -- 错误计数
    
    -- 时间戳
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME,         -- 最后使用时间
    
    CONSTRAINT chk_proxy_type CHECK (proxy_type IN ('socks5', 'http', 'none', NULL))
);

-- 索引
CREATE INDEX idx_claude_accounts_active ON claude_accounts(is_active);
CREATE INDEX idx_claude_accounts_shared ON claude_accounts(is_shared);
CREATE INDEX idx_claude_accounts_rate_limited ON claude_accounts(is_rate_limited);
```

### 3. 管理员会话表 (admin_sessions)

替代Redis的 `session:*` 键

```sql
CREATE TABLE admin_sessions (
    session_id TEXT PRIMARY KEY,   -- 会话ID
    username TEXT NOT NULL,        -- 管理员用户名
    login_time DATETIME NOT NULL,  -- 登录时间
    last_activity DATETIME NOT NULL, -- 最后活动时间
    expires_at DATETIME NOT NULL,  -- 过期时间
    ip_address TEXT,               -- 登录IP
    user_agent TEXT,               -- 用户代理
    is_active BOOLEAN DEFAULT TRUE, -- 是否活跃
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_sessions_username ON admin_sessions(username);
CREATE INDEX idx_sessions_expires ON admin_sessions(expires_at);
CREATE INDEX idx_sessions_active ON admin_sessions(is_active);
```

### 4. 管理员表 (admin_users)

替代Redis的 `admin_credentials` 键

```sql
CREATE TABLE admin_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL, -- 用户名
    password_hash TEXT NOT NULL,   -- bcrypt哈希密码
    email TEXT,                    -- 邮箱（可选）
    is_active BOOLEAN DEFAULT TRUE, -- 是否激活
    last_login_at DATETIME,        -- 最后登录时间
    login_count INTEGER DEFAULT 0, -- 登录次数
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE UNIQUE INDEX idx_admin_users_username ON admin_users(username);
```

### 5. 使用统计表 (usage_records)

替代Redis的 `usage:*` 键

```sql
CREATE TABLE usage_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key_id TEXT NOT NULL,      -- API Key ID
    claude_account_id TEXT,        -- 使用的Claude账户ID
    model TEXT NOT NULL,           -- 模型名称
    token_count INTEGER DEFAULT 0, -- token数量
    cost DECIMAL(10,6) DEFAULT 0,  -- 成本
    request_time DATETIME NOT NULL, -- 请求时间
    response_time_ms INTEGER,      -- 响应时间(毫秒)
    status_code INTEGER,           -- HTTP状态码
    error_message TEXT,            -- 错误信息
    
    -- 按日期分区的辅助字段
    usage_date DATE GENERATED ALWAYS AS (DATE(request_time)) STORED,
    
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id),
    FOREIGN KEY (claude_account_id) REFERENCES claude_accounts(id)
);

-- 索引（优化查询性能）
CREATE INDEX idx_usage_api_key_date ON usage_records(api_key_id, usage_date);
CREATE INDEX idx_usage_account_date ON usage_records(claude_account_id, usage_date);
CREATE INDEX idx_usage_date_model ON usage_records(usage_date, model);
CREATE INDEX idx_usage_request_time ON usage_records(request_time);
```

### 6. 会话映射表 (session_mappings)

替代Redis的sticky会话映射

```sql
CREATE TABLE session_mappings (
    session_hash TEXT PRIMARY KEY, -- 会话哈希
    api_key_id TEXT NOT NULL,      -- API Key ID
    claude_account_id TEXT NOT NULL, -- 绑定的Claude账户ID
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,   -- 过期时间
    
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id),
    FOREIGN KEY (claude_account_id) REFERENCES claude_accounts(id)
);

-- 索引
CREATE INDEX idx_session_mappings_api_key ON session_mappings(api_key_id);
CREATE INDEX idx_session_mappings_account ON session_mappings(claude_account_id);
CREATE INDEX idx_session_mappings_expires ON session_mappings(expires_at);
```

### 7. 系统配置表 (system_config)

替代Redis的系统配置存储

```sql
CREATE TABLE system_config (
    key TEXT PRIMARY KEY,          -- 配置键
    value TEXT,                    -- 配置值（JSON格式）
    description TEXT,              -- 配置描述
    config_type TEXT DEFAULT 'string', -- 配置类型
    is_encrypted BOOLEAN DEFAULT FALSE, -- 是否加密存储
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_config_type CHECK (config_type IN ('string', 'json', 'number', 'boolean'))
);

-- 索引
CREATE INDEX idx_system_config_type ON system_config(config_type);
```

### 8. 限流记录表 (rate_limit_records)

替代Redis的限流存储

```sql
CREATE TABLE rate_limit_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    identifier TEXT NOT NULL,      -- 限流标识符（API Key ID等）
    limit_type TEXT NOT NULL,      -- 限流类型：request/token/concurrency
    current_count INTEGER DEFAULT 0, -- 当前计数
    max_limit INTEGER NOT NULL,    -- 最大限制
    window_start DATETIME NOT NULL, -- 时间窗口开始
    window_duration INTEGER NOT NULL, -- 窗口持续时间（秒）
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_limit_type CHECK (limit_type IN ('request', 'token', 'concurrency'))
);

-- 索引
CREATE UNIQUE INDEX idx_rate_limit_unique ON rate_limit_records(identifier, limit_type);
CREATE INDEX idx_rate_limit_window ON rate_limit_records(window_start);
```

## 数据关系图

```
api_keys (1) ---------> (0..1) claude_accounts (专属绑定)
api_keys (1) ---------> (0..*) usage_records
claude_accounts (1) --> (0..*) usage_records
api_keys (1) ---------> (0..*) session_mappings
claude_accounts (1) --> (0..*) session_mappings
admin_users (1) ------> (0..*) admin_sessions
```

## GORM模型定义

### ApiKey模型

```go
type ApiKey struct {
    ID                   string    `gorm:"primaryKey" json:"id"`
    Name                 string    `gorm:"not null" json:"name"`
    ApiKeyHash          string    `gorm:"uniqueIndex;not null" json:"-"`
    IsActive            bool      `gorm:"default:true" json:"is_active"`
    TokenLimit          int64     `gorm:"default:1000000" json:"token_limit"`
    ConcurrencyLimit    int       `gorm:"default:0" json:"concurrency_limit"`
    CurrentConcurrency  int       `gorm:"default:0" json:"current_concurrency"`
    DedicatedAccountID  *string   `json:"dedicated_account_id"`
    Description         string    `gorm:"default:''" json:"description"`
    CreatedAt           time.Time `json:"created_at"`
    UpdatedAt           time.Time `json:"updated_at"`
    ExpiresAt           *time.Time `json:"expires_at"`
    LastUsedAt          *time.Time `json:"last_used_at"`
    
    // 关联
    DedicatedAccount    *ClaudeAccount `gorm:"foreignKey:DedicatedAccountID" json:"dedicated_account,omitempty"`
    UsageRecords        []UsageRecord  `gorm:"foreignKey:ApiKeyID" json:"-"`
    SessionMappings     []SessionMapping `gorm:"foreignKey:ApiKeyID" json:"-"`
}
```

### ClaudeAccount模型

```go
type ClaudeAccount struct {
    ID                      string    `gorm:"primaryKey" json:"id"`
    Name                    string    `gorm:"not null" json:"name"`
    Email                   string    `json:"email"`
    IsActive               bool      `gorm:"default:true" json:"is_active"`
    IsShared               bool      `gorm:"default:true" json:"is_shared"`
    
    // OAuth数据（加密存储）
    EncryptedOAuthData     string    `json:"-"`
    OAuthScopes            string    `json:"oauth_scopes"`
    OAuthExpiresAt         *time.Time `json:"oauth_expires_at"`
    
    // 代理配置
    ProxyType              string    `json:"proxy_type"`
    ProxyHost              string    `json:"proxy_host"`
    ProxyPort              int       `json:"proxy_port"`
    ProxyUsername          string    `json:"proxy_username"`
    ProxyPasswordEncrypted string    `json:"-"`
    
    // 状态管理
    IsRateLimited          bool      `gorm:"default:false" json:"is_rate_limited"`
    RateLimitUntil         *time.Time `json:"rate_limit_until"`
    LastError              string    `json:"last_error"`
    ErrorCount             int       `gorm:"default:0" json:"error_count"`
    
    // 时间戳
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
    LastUsedAt             *time.Time `json:"last_used_at"`
    
    // 关联
    UsageRecords           []UsageRecord `gorm:"foreignKey:ClaudeAccountID" json:"-"`
    SessionMappings        []SessionMapping `gorm:"foreignKey:ClaudeAccountID" json:"-"`
}
```

## 数据迁移策略

### 1. Redis到SQLite数据迁移工具

```go
type DataMigrator struct {
    redisClient *redis.Client
    db          *gorm.DB
}

func (m *DataMigrator) MigrateApiKeys() error {
    // 扫描Redis中的所有API Key
    // 解析数据并插入到SQLite
}

func (m *DataMigrator) MigrateClaudeAccounts() error {
    // 迁移Claude账户数据
    // 保持加密数据的一致性
}
```

### 2. 增量同步机制

在迁移期间支持数据的增量同步，确保零停机迁移。

## 性能优化

### 1. 索引策略
- 所有外键自动创建索引
- 查询频繁的字段创建复合索引
- 时间范围查询优化

### 2. 分区策略
- usage_records表按月分区
- 历史数据定期归档

### 3. 缓存策略
- 热点数据内存缓存
- API Key验证结果缓存
- 查询结果缓存

## 数据一致性保证

### 1. 事务处理
使用GORM事务确保数据一致性：

```go
func (s *ApiKeyService) CreateApiKeyWithUsage(apiKey *ApiKey, initialUsage *UsageRecord) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(apiKey).Error; err != nil {
            return err
        }
        initialUsage.ApiKeyID = apiKey.ID
        return tx.Create(initialUsage).Error
    })
}
```

### 2. 约束和验证
- 外键约束确保引用完整性
- 检查约束验证数据有效性
- 唯一约束防止重复数据

### 3. 并发控制
- 乐观锁处理并发更新
- 行级锁处理关键操作
- 死锁检测和重试机制

此设计确保了从Redis到SQLite的平滑迁移，同时提供了更好的数据结构化和查询能力。