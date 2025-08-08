# fsd-server

一个用于模拟飞行联飞的FSD, 使用Go语言编写
FSD支持计划同步, 计划锁定, 网页计划提交

[![GitHub release](https://img.shields.io/github/v/release/Flyleague-Collection/fsd-server)](https://www.github.com/Flyleague-Collection/fsd-server/releases/latest)
![GitHub commits since latest release](https://img.shields.io/github/commits-since/Flyleague-Collection/fsd-server/latest/main)
![GitHub top language](https://img.shields.io/github/languages/top/Flyleague-Collection/fsd-server)
![MIT](https://img.shields.io/badge/License-MIT-blue)

## 如何使用

您可以

1. 在[Release](https://www.github.com/Flyleague-Collection/fsd-server/releases/latest)里面下载对应版本的构建
2. 克隆整个存储库自己构建

### 如何构建

1. 克隆本仓库
2. 确保安装了go编译器并且版本>=1.23.4
3. 在项目根目录运行如下命令`go build -x .\cmd\fsd-server\`
4. \[可选\]使用upx压缩可执行文件(windows)`upx.exe -9 .\fsd-server.exe`或者(linux)`upx.exe -9 .\fsd-server`
5. 等待编译完成后, 对于Windows用户, 运行生成的fsd-server.exe; 对于linux用户, 运行生成的fsd-server文件
6. 首次运行会在可执行文件同目录创建配置文件`config.json`, 请在编辑配置文件后再次启动
7. Enjoy

### 配置文件简介
```json5
{
  // 调试模式, 打开后会有大量日志输出, 请不要在生产环境打开
  "debug_mode": false,
  // 服务器名称, 在客户端初次建立连接后会被发送到客户端
  "app_name": "",
  // 服务器版本, 在客户端初次建立连接后会被发送到客户端
  "app_version": "",
  // 最大并发线程数, 也可以看做最大客户端连接数, 推荐值: 64-256
  "max_workers": 0,
  // 最大广播线程数, 推荐与最大并发线程数相同
  // 即每个客户端都可以有一个专门的广播线程, 推荐值: 64-256
  "max_broadcast_workers": 0,
  // 服务器保留客户端连接信息的时间
  // 在这个时间之内客户端重连可以保留之前的数据
  // 反之则清理数据, 推荐值: 40s
  "session_clean_time": "",
  // fsd 服务器配置
  "server_config": {
    // 服务器监听地址, 推荐值: 0.0.0.0
    "host": "",
    // 服务器监听端口, 推荐值: 6809
    "port": 0,
    // 是否启用grpc, 推荐值: true
    "enable_grpc": false,
    // grpc端口, 推荐值: 6810
    // 注: grpc监听地址同服务器监听地址
    "grpc_port": 0,
    // grpc 缓存过期时间, 推荐值: 15s
    "grpc_cache_time": "",
    // 服务器心跳包间隔, 推荐值: 60s
    "heartbeat_interval": "",
    // 服务器motd
    "motd": null
  },
  // 数据库配置
  "database_config": {
    // 数据库地址
    "host": "",
    // 数据库端口
    "port": 0,
    // 数据库用户名
    "username": "",
    // 数据库密码
    "password": "",
    // 数据库名称
    "database": "",
    // 连接空闲时间, 推荐值: 1h
    "connect_idle_timeout": "",
    // 连接超时时间, 推荐值: 5s
    "connect_timeout": "",
    // 数据库最大可用连接, 推荐值: 128
    "server_max_connections": 0
  },
  // 额外权限配置, 详情请看`额外权限配置`章节
  "rating_config": {}
}
```

## 开源协议

MIT License  
Copyright (c) 2025-2025 Half_nothing  
无附加条款。