# Deployment Guide

[English](DEPLOYMENT.md) | [简体中文](DEPLOYMENT.zh-CN.md)

## Overview

This guide covers deploying OpenHost in production environments. OpenHost is designed to be deployed as a containerized application with PostgreSQL and Redis backing services.

## Prerequisites

### System Requirements

**Minimum:**
- CPU: 2 cores
- RAM: 2 GB
- Disk: 20 GB SSD
- OS: Linux (Ubuntu 20.04+ recommended)

**Recommended:**
- CPU: 4+ cores
- RAM: 4+ GB
- Disk: 50+ GB SSD
- OS: Linux (Ubuntu 22.04 LTS)

### Software Dependencies

- Docker 20.10+ and Docker Compose 2.0+
- PostgreSQL 12+ (or use Docker)
- Redis 6+ (or use Docker)
- Nginx or Traefik (for reverse proxy)
- SSL certificates (Let's Encrypt recommended)

## Deployment Options

### Option 1: Docker Compose (Recommended)

This is the simplest way to deploy OpenHost with all dependencies.

#### 1. Create docker-compose.yml

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

#### 2. Create .env file

```bash
# Database
DB_PASSWORD=your_secure_db_password_here

# Redis
REDIS_PASSWORD=your_secure_redis_password_here

# Application
SERVER_PORT=8080
JWT_SECRET=your_jwt_secret_key_here
```

#### 3. Deploy

```bash
# Pull images
docker-compose pull

# Start services
docker-compose up -d

# Check logs
docker-compose logs -f openhost

# Check status
docker-compose ps
```

#### 4. Initialize Database

```bash
# Run migrations
docker-compose exec openhost /app/bin/migrate

# Or manually
docker-compose exec postgres psql -U openhost -d openhost -f /sql/schema.sql
```

### Option 2: Kubernetes

For production-scale deployments with high availability.

#### 1. Create ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: openhost-config
  namespace: openhost
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "openhost"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
```

#### 2. Create Secrets

```bash
kubectl create secret generic openhost-secrets \
  --from-literal=db-password='your_secure_password' \
  --from-literal=redis-password='your_redis_password' \
  --from-literal=jwt-secret='your_jwt_secret' \
  -n openhost
```

#### 3. Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openhost
  namespace: openhost
spec:
  replicas: 3
  selector:
    matchLabels:
      app: openhost
  template:
    metadata:
      labels:
        app: openhost
    spec:
      containers:
      - name: openhost
        image: openhost/openhost:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: openhost-config
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: openhost-secrets
              key: db-password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: openhost-secrets
              key: redis-password
        volumeMounts:
        - name: plugins
          mountPath: /app/plugins
        - name: themes
          mountPath: /app/themes
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: plugins
        persistentVolumeClaim:
          claimName: openhost-plugins-pvc
      - name: themes
        persistentVolumeClaim:
          claimName: openhost-themes-pvc
```

#### 4. Create Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: openhost-service
  namespace: openhost
spec:
  selector:
    app: openhost
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### Option 3: Manual Installation

For bare metal or VM deployments.

#### 1. Install Dependencies

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# Install Redis
sudo apt install redis-server -y

# Install Nginx
sudo apt install nginx -y
```

#### 2. Setup Database

```bash
# Connect to PostgreSQL
sudo -u postgres psql

# Create database and user
CREATE DATABASE openhost;
CREATE USER openhost WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE openhost TO openhost;
\q
```

#### 3. Configure Redis

```bash
# Edit Redis config
sudo nano /etc/redis/redis.conf

# Set password
requirepass your_redis_password

# Restart Redis
sudo systemctl restart redis
```

#### 4. Build and Install OpenHost

```bash
# Clone repository
git clone https://github.com/lbyxiaolizi/xlpanel.git
cd xlpanel

# Build
make all

# Install binaries
sudo mkdir -p /opt/openhost
sudo cp -r bin /opt/openhost/
sudo cp -r themes /opt/openhost/
sudo mkdir -p /opt/openhost/plugins

# Create systemd service
sudo nano /etc/systemd/system/openhost.service
```

#### 5. Create Systemd Service

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

#### 6. Start Service

```bash
# Create user
sudo useradd -r -s /bin/false openhost
sudo chown -R openhost:openhost /opt/openhost

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable openhost
sudo systemctl start openhost

# Check status
sudo systemctl status openhost
```

## Reverse Proxy Configuration

### Nginx

```nginx
upstream openhost {
    server localhost:8080;
}

server {
    listen 80;
    server_name yourdomain.com;
    
    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    # Proxy configuration
    location / {
        proxy_pass http://openhost;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static files
    location /static/ {
        alias /opt/openhost/themes/default/assets/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # File upload limits
    client_max_body_size 100M;
}
```

### Traefik (Docker)

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    command:
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@yourdomain.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./letsencrypt:/letsencrypt
    networks:
      - openhost

  openhost:
    image: openhost/openhost:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.openhost.rule=Host(`yourdomain.com`)"
      - "traefik.http.routers.openhost.entrypoints=websecure"
      - "traefik.http.routers.openhost.tls.certresolver=letsencrypt"
    networks:
      - openhost
```

## Database Management

### Backup

```bash
# Manual backup
pg_dump -U openhost openhost > backup_$(date +%Y%m%d_%H%M%S).sql

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backups/openhost"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
pg_dump -U openhost openhost | gzip > $BACKUP_DIR/backup_$DATE.sql.gz
find $BACKUP_DIR -mtime +7 -delete  # Keep 7 days
```

### Restore

```bash
# Restore from backup
psql -U openhost openhost < backup_20240101_120000.sql

# Or from gzip
gunzip -c backup_20240101_120000.sql.gz | psql -U openhost openhost
```

### Migrations

```bash
# Run pending migrations
./bin/server --migrate

# Rollback last migration
./bin/server --migrate-rollback

# Check migration status
./bin/server --migrate-status
```

## Monitoring

### Health Checks

```bash
# Application health
curl http://localhost:8080/api/v1/health

# Expected response
{"status":"ok"}
```

### Logs

```bash
# Docker Compose
docker-compose logs -f openhost

# Systemd
sudo journalctl -u openhost -f

# Application logs
tail -f /var/log/openhost/app.log
```

### Metrics

OpenHost exposes Prometheus-compatible metrics at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

## Security Checklist

- [ ] Use strong passwords for database and Redis
- [ ] Enable SSL/TLS for all connections
- [ ] Configure firewall (UFW/iptables)
- [ ] Set up fail2ban for brute force protection
- [ ] Regular security updates
- [ ] Limit SSH access (key-based auth only)
- [ ] Configure proper file permissions
- [ ] Enable audit logging
- [ ] Regular backups to secure location
- [ ] Monitor security logs

## Performance Tuning

### Database

```sql
-- Increase connection pool
ALTER SYSTEM SET max_connections = 200;

-- Adjust shared buffers (25% of RAM)
ALTER SYSTEM SET shared_buffers = '1GB';

-- Enable query planning
ALTER SYSTEM SET effective_cache_size = '3GB';

-- Reload configuration
SELECT pg_reload_conf();
```

### Redis

```bash
# In redis.conf
maxmemory 512mb
maxmemory-policy allkeys-lru
```

### Application

```bash
# Environment variables
export GOMAXPROCS=4  # Number of CPU cores
export GOMEMLIMIT=2GiB  # Memory limit
```

## Troubleshooting

### Application Won't Start

```bash
# Check logs
docker-compose logs openhost

# Check database connection
docker-compose exec openhost psql -h postgres -U openhost

# Verify environment variables
docker-compose exec openhost env | grep DB_
```

### High Memory Usage

```bash
# Check container stats
docker stats

# Restart service
docker-compose restart openhost

# Adjust memory limits in docker-compose.yml
```

### Slow Database Queries

```sql
-- Enable slow query log
ALTER SYSTEM SET log_min_duration_statement = 1000;  -- 1 second

-- Check slow queries
SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;

-- Analyze tables
ANALYZE;
```

## Scaling

### Horizontal Scaling

1. Use load balancer (Nginx/HAProxy)
2. Deploy multiple OpenHost instances
3. Shared PostgreSQL and Redis
4. Shared file storage (NFS/S3)

### Vertical Scaling

1. Increase CPU/RAM
2. Optimize database indexes
3. Enable caching
4. Use connection pooling

## Updates

### Docker Compose

```bash
# Pull latest image
docker-compose pull openhost

# Recreate containers
docker-compose up -d --force-recreate

# Clean old images
docker image prune -a
```

### Manual Installation

```bash
# Backup current installation
sudo cp -r /opt/openhost /opt/openhost.backup

# Stop service
sudo systemctl stop openhost

# Update binaries
cd xlpanel
git pull
make all
sudo cp -r bin/* /opt/openhost/bin/

# Start service
sudo systemctl start openhost

# Check logs
sudo journalctl -u openhost -f
```

## Support

For deployment issues:
- Check [GitHub Issues](https://github.com/lbyxiaolizi/xlpanel/issues)
- Read [Architecture Documentation](ARCHITECTURE.md)
- Join our community forum

## Additional Resources

- [PostgreSQL Tuning Guide](https://wiki.postgresql.org/wiki/Tuning_Your_PostgreSQL_Server)
- [Redis Best Practices](https://redis.io/topics/admin)
- [Nginx Configuration](https://nginx.org/en/docs/)
- [Docker Production Guide](https://docs.docker.com/config/daemon/)
