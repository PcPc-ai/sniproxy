> ## 🚀 [PcPc.AI](https://www.pcpc.ai) — 一个账号畅享 GPT / Gemini / Claude 最新模型
> AI 生图、文字创作，**无额度限制**！👉 https://www.pcpc.ai
>
> *One account for the latest GPT, Gemini & Claude models — AI image generation and text creation with no quota limits.*

---

# SNI Proxy Server

[English](#english) | [中文](#中文)

---

## English

A Go-based SNI (Server Name Indication) proxy server that listens for HTTPS and HTTP traffic, forwards connections to the target host parsed from the TLS ClientHello, and supports suffix/exact-match domain whitelisting.

### Features

- **HTTPS proxy**: parses SNI from the TLS ClientHello and transparently proxies the connection.
- **HTTP proxy**: forwards HTTP requests based on the `Host` header.
- **Whitelist filtering**: supports both exact-match and suffix-match domain rules.
- **Config file**: YAML configuration.
- **Flexible deployment**: configurable listen ports and bind address.
- **Rock-solid stability**: no OOM, no FD leaks; runs continuously for 180+ days on a 1 vCPU / 1 GB RAM machine.

### Quick Start

#### 1. Build

```bash
go build -o sniproxy .
```

#### 2. Configure the whitelist

Edit `domains.csv` and add the allowed domain rules:

```csv
openai.com,suffix
chatgpt.com,suffix
google.com,suffix
example.com,exact
```

Rule format: `domain,type`
- `suffix`: suffix match (allows subdomains).
- `exact`: exact match.

#### 3. Edit the config

Edit `config.yaml`:

```yaml
server:
  https_port: "443"
  http_port: "80"
  bind_addr: "0.0.0.0"

whitelist:
  enabled: true
  default_mode: "deny"
  domains_file: "domains.csv"
  domains:
    - "example.com"

logging:
  level: "info"
  format: "text"
```

#### 4. Run

Using the start script:

```bash
./start.sh
```

Or run the binary directly:

```bash
./sniproxy -config config.yaml
```

### Configuration Reference

#### `server`

- `https_port`: HTTPS listen port (default `443`).
- `http_port`: HTTP listen port (default `80`).
- `bind_addr`: bind address (default `0.0.0.0`).

#### `whitelist`

- `enabled`: enable whitelist filtering.
- `default_mode`: default action when no rule matches — `allow` or `deny`.
- `domains_file`: path to the domain rules file.
- `domains`: inline domain list (treated as suffix match).

#### `logging`

- `level`: `debug`, `info`, `warn`, or `error`.
- `format`: `text` or `json`.

### Whitelist Rules

1. **exact** — `example.com,exact` matches only `example.com`.
2. **suffix** — `example.com,suffix` matches `example.com`, `www.example.com`, `api.example.com`, etc.

Rules file format (e.g. `domains.csv`):

```csv
# Lines starting with # are comments
domain.com,suffix
exact-domain.org,exact
```

### Permissions

Binding to privileged ports (80/443) requires elevated privileges:

1. **Run as root**:
   ```bash
   sudo ./sniproxy -config config.yaml
   ```

2. **Grant `CAP_NET_BIND_SERVICE`**:
   ```bash
   sudo setcap CAP_NET_BIND_SERVICE=+eip ./sniproxy
   ./sniproxy -config config.yaml
   ```

3. **Use non-privileged ports**: change the ports in `config.yaml` to values `> 1024`.

### How It Works

**HTTPS flow**

1. Accept the client connection on the HTTPS port.
2. Read the TLS ClientHello.
3. Parse the SNI extension to obtain the target hostname.
4. Check the whitelist.
5. Dial the target server.
6. Forward the original ClientHello to the target.
7. Start bidirectional relay.

**HTTP flow**

1. Accept the client connection on the HTTP port.
2. Parse the HTTP request and read the `Host` header.
3. Check the whitelist.
4. Dial the target server.
5. Forward the HTTP request.
6. Start bidirectional relay.

### Project Layout

```
sniproxy/
├── main.go              # Entry point
├── config.yaml          # Config file
├── domains.csv          # Whitelist rules
├── start.sh             # Start script
├── internal/
│   ├── config/          # Config loading
│   ├── proxy/           # Proxy core (server / https / http)
│   ├── sni/             # SNI parser
│   ├── whitelist/       # Whitelist manager
│   └── logger/          # Logging
└── go.mod
```

### Notes

1. Make sure the target server is reachable.
2. The proxy does not decrypt or modify HTTPS payloads.
3. The HTTP proxy parses request headers but does not modify the body.
4. Tune the log level appropriately in production.
5. Keep the whitelist updated.

---

## 中文

一个基于Go语言开发的SNI（Server Name Indication）代理服务器，能够监听HTTPS和HTTP流量，根据TLS包中的SNI信息将连接转发到对应的域名，同时支持基于域名后缀的白名单规则配置。

### 功能特性

- **HTTPS代理**: 解析TLS ClientHello中的SNI信息，透明代理HTTPS连接
- **HTTP代理**: 支持HTTP请求代理，基于Host头进行转发
- **白名单过滤**: 支持精确匹配和后缀匹配的域名白名单规则
- **配置文件**: 支持YAML格式的配置文件
- **灵活部署**: 支持自定义监听端口和绑定地址
- **稳定性优秀**: 无OOM、无FD泄漏，在1C 1G配置的机器上可稳定运行超过180天

### 快速开始

#### 1. 编译项目

```bash
go build -o sniproxy .
```

#### 2. 配置白名单

编辑 `domains.csv` 文件，添加允许的域名规则：

```csv
openai.com,suffix
chatgpt.com,suffix
google.com,suffix
example.com,exact
```

规则格式：`域名,类型`
- `suffix`: 后缀匹配（允许子域名）
- `exact`: 精确匹配

#### 3. 修改配置

编辑 `config.yaml` 文件：

```yaml
server:
  https_port: "443"
  http_port: "80"
  bind_addr: "0.0.0.0"

whitelist:
  enabled: true
  default_mode: "deny"
  domains_file: "domains.csv"
  domains:
    - "example.com"

logging:
  level: "info"
  format: "text"
```

#### 4. 启动服务器

使用启动脚本：

```bash
./start.sh
```

或直接运行：

```bash
./sniproxy -config config.yaml
```

### 配置说明

#### 服务器配置

- `https_port`: HTTPS监听端口（默认443）
- `http_port`: HTTP监听端口（默认80）
- `bind_addr`: 绑定地址（默认0.0.0.0）

#### 白名单配置

- `enabled`: 是否启用白名单过滤
- `default_mode`: 默认模式，`allow`（允许所有）或`deny`（拒绝所有）
- `domains_file`: 域名规则文件路径
- `domains`: 配置文件中直接指定的域名列表（后缀匹配）

#### 日志配置

- `level`: 日志级别（debug, info, warn, error）
- `format`: 日志格式（text, json）

### 白名单规则

#### 规则类型

1. **exact**: 精确匹配
   - `example.com,exact` 只匹配 `example.com`

2. **suffix**: 后缀匹配
   - `example.com,suffix` 匹配 `example.com`, `www.example.com`, `api.example.com` 等

#### 规则文件格式

域名规则文件（如`domains.csv`）格式：

```csv
# 注释行以#开头
domain.com,suffix
exact-domain.org,exact
```

### 权限要求

绑定特权端口（80/443）需要相应权限：

1. **使用root权限运行**:
   ```bash
   sudo ./sniproxy -config config.yaml
   ```

2. **授予CAP_NET_BIND_SERVICE权限**:
   ```bash
   sudo setcap CAP_NET_BIND_SERVICE=+eip ./sniproxy
   ./sniproxy -config config.yaml
   ```

3. **使用非特权端口**:
   修改配置文件中的端口为>1024的端口

### 工作原理

#### HTTPS代理流程

1. 监听HTTPS端口接收客户端连接
2. 读取TLS ClientHello握手包
3. 解析SNI扩展获取目标主机名
4. 检查主机名是否在白名单中
5. 建立到目标服务器的连接
6. 转发ClientHello包到目标服务器
7. 开始双向数据转发

#### HTTP代理流程

1. 监听HTTP端口接收客户端连接
2. 解析HTTP请求获取Host头
3. 检查主机名是否在白名单中
4. 建立到目标服务器的连接
5. 转发HTTP请求到目标服务器
6. 开始双向数据转发

### 项目结构

```
sniproxy/
├── main.go                 # 主程序入口
├── config.yaml            # 配置文件
├── domains.csv            # 域名白名单文件
├── start.sh               # 启动脚本
├── internal/
│   ├── config/            # 配置管理
│   ├── proxy/             # 代理核心逻辑
│   ├── sni/               # SNI解析
│   ├── whitelist/         # 白名单管理
│   └── logger/            # 日志
└── go.mod                 # Go模块文件
```

### 注意事项

1. 确保目标服务器可达
2. 代理不会修改或解密HTTPS流量内容
3. HTTP代理会解析请求头但不会修改内容
4. 建议在生产环境中配置适当的日志级别
5. 定期更新白名单规则以确保安全性
