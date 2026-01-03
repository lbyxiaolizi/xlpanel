# OpenHost 安装指南

本指南介绍如何使用内置 SQLite（推荐快速体验）或 PostgreSQL 安装 OpenHost。

## 环境要求

- Go 1.22+（推荐 Go 1.23 toolchain）
- GCC 编译工具链（`go-sqlite3` 需要）
- 可选：PostgreSQL 13+

## 构建服务端

```bash
make server
```

二进制文件输出到 `./bin/server`。

## 启动服务

```bash
./bin/server
```

浏览器访问：

```
http://localhost:6421/install
```

## 安装向导步骤

1. **站点设置**：填写站点名称与基础地址。
2. **管理员账户**：填写管理员邮箱与密码（密码以 bcrypt 哈希保存）。
3. **数据库**：
   - **SQLite（内置）**：默认路径 `./data/openhost.db`。
   - **PostgreSQL**：填写主机、端口、用户名、密码、数据库名称。

提交后安装程序将：

- 创建 SQLite 数据库文件或连接 PostgreSQL。
- 运行核心领域表的 GORM 自动迁移。
- 写入配置文件 `config/openhost.json`。

## 重新安装

如需重新安装：

1. 停止服务。
2. 删除 `config/openhost.json`。
3. 若使用 SQLite，删除数据库文件，例如 `./data/openhost.db`。
4. 启动服务并重新访问 `/install`。

## 说明

- 配置文件权限为 `0600`。
- SQLite 默认存储在 `./data` 目录。
- PostgreSQL 用户需具备建表权限。
