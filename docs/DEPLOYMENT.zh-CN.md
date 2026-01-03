# 部署指南

[English](DEPLOYMENT.md) | 简体中文

## 概述

本指南介绍如何在生产环境中部署 OpenHost。OpenHost 设计为容器化应用程序，使用 PostgreSQL 和 Redis 作为后端服务。

## 前置要求

### 系统要求

**最低配置：**
- CPU: 2 核心
- 内存: 2 GB
- 磁盘: 20 GB SSD
- 操作系统: Linux (推荐 Ubuntu 20.04+)

**推荐配置：**
- CPU: 4+ 核心
- 内存: 4+ GB
- 磁盘: 50+ GB SSD
- 操作系统: Linux (Ubuntu 22.04 LTS)

### 软件依赖

- Docker 20.10+ 和 Docker Compose 2.0+
- PostgreSQL 12+ (或使用 Docker)
- Redis 6+ (或使用 Docker)
- Nginx 或 Traefik (用于反向代理)
- SSL 证书 (推荐 Let's Encrypt)

## 部署选项

### 选项 1: Docker Compose (推荐)

这是部署 OpenHost 及其所有依赖项的最简单方式。

#### 1. 创建 docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: openhost
      POSTGRES_USER: openhost
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - openhost
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    networks:
      - openhost
    restart: unless-stopped

  openhost:
    image: openhost/openhost:latest
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=openhost
      - DB_USER=openhost
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - SERVER_PORT=8080
    ports:
      - "8080:8080"
    volumes:
      - ./plugins:/app/plugins
      - ./themes:/app/themes
      - ./uploads:/app/uploads
    networks:
      - openhost
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:

networks:
  openhost:
```

#### 2. 创建 .env 文件

```bash
# 数据库
DB_PASSWORD=your_secure_db_password_here

# Redis
REDIS_PASSWORD=your_secure_redis_password_here

# 应用程序
SERVER_PORT=8080
JWT_SECRET=your_jwt_secret_key_here
```

#### 3. 部署

```bash
# 拉取镜像
docker-compose pull

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f openhost

# 检查状态
docker-compose ps
```

#### 4. 初始化数据库

```bash
# 运行迁移
docker-compose exec openhost /app/bin/migrate

# 或手动执行
docker-compose exec postgres psql -U openhost -d openhost -f /sql/schema.sql
```

### 选项 2: 手动安装

用于裸机或虚拟机部署。

#### 1. 安装依赖

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# 安装 Redis
sudo apt install redis-server -y

# 安装 Nginx
sudo apt install nginx -y
```

#### 2. 设置数据库

```bash
# 连接到 PostgreSQL
sudo -u postgres psql

# 创建数据库和用户
CREATE DATABASE openhost;
CREATE USER openhost WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE openhost TO openhost;
\q
```

#### 3. 配置 Redis

```bash
# 编辑 Redis 配置
sudo nano /etc/redis/redis.conf

# 设置密码
requirepass your_redis_password

# 重启 Redis
sudo systemctl restart redis
```

#### 4. 构建和安装 OpenHost

```bash
# 克隆仓库
git clone https://github.com/lbyxiaolizi/xlpanel.git
cd xlpanel

# 构建
make all

# 安装二进制文件
sudo mkdir -p /opt/openhost
sudo cp -r bin /opt/openhost/
sudo cp -r themes /opt/openhost/
sudo mkdir -p /opt/openhost/plugins
```

#### 5. 创建 Systemd 服务

```bash
sudo nano /etc/systemd/system/openhost.service
```

```ini
[Unit]
Description=OpenHost Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=openhost
Group=openhost
WorkingDirectory=/opt/openhost
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_NAME=openhost"
Environment="DB_USER=openhost"
Environment="DB_PASSWORD=your_password"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"
Environment="REDIS_PASSWORD=your_redis_password"
ExecStart=/opt/openhost/bin/server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

#### 6. 启动服务

```bash
# 创建用户
sudo useradd -r -s /bin/false openhost
sudo chown -R openhost:openhost /opt/openhost

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable openhost
sudo systemctl start openhost

# 检查状态
sudo systemctl status openhost
```

## 反向代理配置

### Nginx

```nginx
upstream openhost {
    server localhost:8080;
}

server {
    listen 80;
    server_name yourdomain.com;
    
    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    # SSL 配置
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 安全头
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    # 代理配置
    location / {
        proxy_pass http://openhost;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 静态文件
    location /static/ {
        alias /opt/openhost/themes/default/assets/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # 文件上传限制
    client_max_body_size 100M;
}
```

## 数据库管理

### 备份

```bash
# 手动备份
pg_dump -U openhost openhost > backup_$(date +%Y%m%d_%H%M%S).sql

# 自动备份脚本
#!/bin/bash
BACKUP_DIR="/backups/openhost"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
pg_dump -U openhost openhost | gzip > $BACKUP_DIR/backup_$DATE.sql.gz
find $BACKUP_DIR -mtime +7 -delete  # 保留 7 天
```

### 恢复

```bash
# 从备份恢复
psql -U openhost openhost < backup_20240101_120000.sql

# 或从 gzip 恢复
gunzip -c backup_20240101_120000.sql.gz | psql -U openhost openhost
```

## 监控

### 健康检查

```bash
# 应用程序健康
curl http://localhost:8080/api/v1/health

# 预期响应
{"status":"ok"}
```

### 日志

```bash
# Docker Compose
docker-compose logs -f openhost

# Systemd
sudo journalctl -u openhost -f

# 应用程序日志
tail -f /var/log/openhost/app.log
```

## 安全检查清单

- [ ] 为数据库和 Redis 使用强密码
- [ ] 为所有连接启用 SSL/TLS
- [ ] 配置防火墙 (UFW/iptables)
- [ ] 设置 fail2ban 防止暴力破解
- [ ] 定期安全更新
- [ ] 限制 SSH 访问 (仅使用密钥认证)
- [ ] 配置适当的文件权限
- [ ] 启用审计日志
- [ ] 定期备份到安全位置
- [ ] 监控安全日志

## 性能调优

### 数据库

```sql
-- 增加连接池
ALTER SYSTEM SET max_connections = 200;

-- 调整共享缓冲区 (RAM 的 25%)
ALTER SYSTEM SET shared_buffers = '1GB';

-- 启用查询规划
ALTER SYSTEM SET effective_cache_size = '3GB';

-- 重新加载配置
SELECT pg_reload_conf();
```

### Redis

```bash
# 在 redis.conf 中
maxmemory 512mb
maxmemory-policy allkeys-lru
```

## 故障排除

### 应用程序无法启动

```bash
# 检查日志
docker-compose logs openhost

# 检查数据库连接
docker-compose exec openhost psql -h postgres -U openhost

# 验证环境变量
docker-compose exec openhost env | grep DB_
```

### 高内存使用

```bash
# 检查容器统计信息
docker stats

# 重启服务
docker-compose restart openhost
```

## 扩展

### 水平扩展

1. 使用负载均衡器 (Nginx/HAProxy)
2. 部署多个 OpenHost 实例
3. 共享 PostgreSQL 和 Redis
4. 共享文件存储 (NFS/S3)

### 垂直扩展

1. 增加 CPU/RAM
2. 优化数据库索引
3. 启用缓存
4. 使用连接池

## 更新

### Docker Compose

```bash
# 拉取最新镜像
docker-compose pull openhost

# 重新创建容器
docker-compose up -d --force-recreate

# 清理旧镜像
docker image prune -a
```

## 支持

有关部署问题：
- 查看 [GitHub Issues](https://github.com/lbyxiaolizi/xlpanel/issues)
- 阅读 [架构文档](ARCHITECTURE.zh-CN.md)
- 加入我们的社区论坛

## 其他资源

- [PostgreSQL 调优指南](https://wiki.postgresql.org/wiki/Tuning_Your_PostgreSQL_Server)
- [Redis 最佳实践](https://redis.io/topics/admin)
- [Nginx 配置](https://nginx.org/en/docs/)
- [Docker 生产指南](https://docs.docker.com/config/daemon/)
