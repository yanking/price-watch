# 授权模块集成说明

## 配置文件

### 1. 主配置 (configs/watch.yaml)

包含日志、MySQL 和 Redis 配置:

```yaml
log:
  level: 'info'
  format: 'json'
  output: 'stdout'
  enableTrace: true
  addSource: false
  timeFormat: '2006-01-02 15:04:05'

mysql:
  dsn: "root:root@tcp(127.0.0.1:3306)/watch?charset=utf8mb4&parseTime=True&loc=Local"
  maxIdleConns: 10
  maxOpenConns: 100
  maxLifetime: 3600

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  poolSize: 10
```

### 2. Auth 配置 (configs/auth.yaml)

包含 JWT 和 OAuth 配置:

```yaml
jwt:
  secret: "your-secret-key-here"
  expire_days: 7

oauth:
  github:
    client_id: "github-client-id"
    client_secret: "github-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/oauth/callback/github"
  wechat:
    client_id: "wechat-client-id"
    client_secret: "wechat-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/oauth/callback/wechat"
```

## 启动服务

```bash
# 使用默认配置启动
go run ./cmd/watch

# 指定配置文件
go run ./cmd/watch -c ./configs/watch.yaml -auth ./configs/auth.yaml
```

## HTTP 服务

服务启动后,HTTP 服务器监听在 `0.0.0.0:8080`

### 健康检查

```bash
curl http://localhost:8080/health
```

### API 路由

认证模块路由将在数据库初始化后启用:

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/oauth` - OAuth 登录
- `GET /api/v1/user/profile` - 获取用户信息 (需要认证)
- `PUT /api/v1/user/profile` - 更新用户信息 (需要认证)
- `PUT /api/v1/user/password` - 修改密码 (需要认证)

## 下一步

1. 创建数据库表
2. 在 HTTP 服务器中集成 auth 路由
3. 配置 OAuth 应用
4. 测试 API 接口
