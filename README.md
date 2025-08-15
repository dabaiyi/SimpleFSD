# fsd-server



一个用于模拟飞行联飞的FSD, 使用Go语言编写

FSD支持计划同步, 计划锁定, 网页计划提交

此项目实际上除FSD以外还包括许多有用的api接口，可以在[此处](#链接)找到说明文档，理论上是：FSD+后端的结合体

---
[![GitHub release](https://img.shields.io/github/v/release/Flyleague-Collection/fsd-server?style=for-the-badge&logo=github)](https://www.github.com/Flyleague-Collection/fsd-server/releases/latest)
![GitHub Release Date](https://img.shields.io/github/release-date/Flyleague-Collection/fsd-server?display_date=published_at&style=for-the-badge&logo=github)  
![GitHub last commit](https://img.shields.io/github/last-commit/Flyleague-Collection/fsd-server?display_timestamp=committer&style=for-the-badge&logo=github)
![Build](https://img.shields.io/github/actions/workflow/status/Flyleague-Collection/fsd-server/go-build.yml?style=for-the-badge&logo=github)  
![GitHub top language](https://img.shields.io/github/languages/top/Flyleague-Collection/fsd-server?style=for-the-badge&logo=github)
![MIT](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge&logo=github)



## 如何使用

### 使用方法

1. 在[Release](https://www.github.com/Flyleague-Collection/fsd-server/releases/latest)里面下载对应版本的构建
2. 克隆整个存储库自己构建



### 构建方法

1. 克隆本仓库
2. 确保安装了go编译器并且版本>=1.23.4
3. 在项目根目录运行如下命令`go build -x .\cmd\fsd-server\`
4. \[可选\]使用upx压缩可执行文件(windows)`upx.exe -9 .\fsd-server.exe`或者(linux)`upx -9 .\fsd-server`
5. 等待编译完成后, 对于Windows用户, 运行生成的fsd-server.exe; 对于linux用户, 运行生成的fsd-server文件。注：默认无进程守护，因此推荐使用，宝塔面板的“网站 - Go项目 - 添加项目”进行使用。
6. 首次运行会在可执行文件同目录创建配置文件`config.json`, 请在编辑配置文件后再次启动
7. Enjoy



### 配置文件简介

```json5
{
  // 调试模式, 会输出大量日志, 请不要在生产环境中打开
  "debug_mode": false,
  // 配置文件版本, 通常情况下与软件版本一致
  "config_version": "0.5.0",
  // 服务配置
  "server": {
    // 通用配置项
    "general": {
      // 是否为模拟机服务器
      // 由于需要实现检查网页提交计划于实际连线计划是否一致
      // 所以飞行计划存储是用用户cid进行标识的
      // 但模拟机所有的模拟机都是一个用户cid, 此时就会出问题
      // 即模拟机计划错误或者无法获取到计划
      // 这个时候将这个变量设置为true
      // 这样服务器就会使用呼号作为标识
      // 但是与此同时就失去了呼号匹配检查的功能
      // 但网页提交计划仍然可用, 只是没有检查功能
      // 所以将这个开关命名为模拟机服务器开关
      "simulator_server": false,
      // 密码加密轮数
      "bcrypt_cost": 12,
      // jwt对称加密秘钥
      "jwt_secret": "123456",
      // jwt秘钥过期时间
      "jwt_expires_time": "1h",
      // jwt刷新秘钥过期时间
      // 该时间是在jwt秘钥过期时间之后的时间
      // 比如两者都是1h, 那么刷新秘钥的过期时间就是2h
      // 因为不可能你刷新秘钥比主密钥过期还早:(
      "jwt_refresh_time": "1h"
    },
    // FSD服务器配置
    "fsd_server": {
      // FSD名称, 会被发送到连接到服务器的客户端作为motd消息
      "fsd_name": "Simple-Fsd",
      // FSD服务器监听地址
      "host": "0.0.0.0",
      // FSD服务器监听端口
      "port": 6809,
      // 是否发送wallop消息到ADM
      "send_wallop_to_adm": true,
      // FSD服务器心跳间隔
      "heartbeat_interval": "60s",
      // FSD服务器会话过期时间
      // 在过期时间内重连, 服务器会自动匹配断开时的session
      // 反之则会创建新session
      "session_clean_time": "40s",
      // 最大工作线程数, 也可以理解为最大同时连接的sockets数目
      "max_workers": 128,
      // 最大广播线程数, 用于广播消息的最大线程数
      "max_broadcast_workers": 128,
      // 要发送到客户端的motd消息
      "motd": []
    },
    // Http服务器配置
    "http_server": {
      // 是否启用Http服务器
      "enabled": false,
      // Http服务器监听地址
      "host": "0.0.0.0",
      // Http服务器监听端口
      "port": 6810,
      // Http服务器最大工作线程
      "max_workers": 128,
      // Http服务器Api缓存时间
      "cache_time": "15s",
      // 是否启用SSL
      "enable_ssl": false,
      // 如果启用SSL, 这里填写证书路径
      "cert_file": "",
      // 如果启用SSL, 这里填写私钥路径
      "key_file": ""
    },
    // gRPC服务器
    "grpc_server": {
      // 是否启用gRPC服务器
      "enabled": false,
      // gRPC服务器监听地址
      "host": "0.0.0.0",
      // gRPC服务器监听端口
      "port": 6811,
      // gRPC服务器Api缓存时间
      "cache_time": "15s"
    }
  },
  // 数据库配置
  "database": {
    // 数据库类型, 支持的数据库类型: mysql, postgres, sqlite3
    "type": "mysql",
    // 当数据库类型为sqlite3的时候, 这里是数据库存放路径和文件名
    // 反之则为要使用的数据库名称
    "database": "go-fsd",
    // 数据库地址
    "host": "localhost",
    // 数据库端口
    "port": 3306,
    // 数据库用户名
    "username": "root",
    // 数据库密码
    "password": "123456",
    // 是否启用SSL
    "enable_ssl": false,
    // 数据库连接池连接超时时间
    "connect_idle_timeout": "1h",
    // 连接超时时间
    "connect_timeout": "5s",
    // 数据库最大连接数
    "server_max_connections": 32
  },
  // 特殊权限配置, 详情请见`特殊权限配置` 章节
  "rating": {}
}
```



### 权限定义表

#### FSD管制权限一览

| 权限识别名         | 权限值 | 中文名   | 说明                      |
|:--------------|:---:|:------|:------------------------|
| Ban           | -1  | 封禁    |                         |
| Normal        |  0  | 普通用户  | 默认权限                    |
| Observer      |  1  | 观察者   |                         |
| STU1          |  2  | 放行/地面 |                         |
| STU2          |  3  | 塔台    |                         |
| STU3          |  4  | 终端    |                         |
| CTR1          |  5  | 区域    |                         |
| CTR2          |  6  | 区域    | 该权限已被弃用, 这里写出来只是为了与ES同步 |
| CTR3          |  7  | 区域    |                         |
| Instructor1   |  8  | 教员    |                         |
| Instructor2   |  9  | 教员    |                         |
| Instructor3   | 10  | 教员    |                         |
| Supervisor    | 11  | 监察者   |                         |
| Administrator | 12  | 管理员   |                         |



#### 管制席位一览

| 席位识别名 | 席位编码 | 中文名 | 说明         |
|:------|:-----|:----|:-----------|
| Pilot | 1    | 飞行员 | 连线飞行员属于该席位 |
| OBS   | 2    | 观察者 |            |
| DEL   | 4    | 放行  |            |
| GND   | 8    | 地面  |            |
| TWR   | 16   | 塔台  |            |
| APP   | 32   | 进近  |            |
| CTR   | 64   | 区域  |            |
| FSS   | 128  | 飞服  |            |



#### 管制权限与管制席位对照一览

| 权限识别名         | 允许的席位                             | 说明   |
|:--------------|:----------------------------------|:-----|
| Ban           | 不允许任何席位                           | 封禁用户 |
| Normal        | Pilot                             |      |
| Observer      | Pilot,OBS                         |      |
| STU1          | Pilot,OBS,DEL,GND                 |      |
| STU2          | Pilot,OBS,DEL,GND,TWR             |      |
| STU3          | Pilot,OBS,DEL,GND,TWR,APP         |      |
| CTR1          | Pilot,OBS,DEL,GND,TWR,APP,CTR     |      |
| CTR2          | Pilot,OBS,DEL,GND,TWR,APP,CTR     |      |
| CTR3          | Pilot,OBS,DEL,GND,TWR,APP,CTR,FSS |      |
| Instructor1   | Pilot,OBS,DEL,GND,TWR,APP,CTR,FSS |      |
| Instructor2   | Pilot,OBS,DEL,GND,TWR,APP,CTR,FSS |      |
| Instructor3   | Pilot,OBS,DEL,GND,TWR,APP,CTR,FSS |      |
| Supervisor    | Pilot,OBS,DEL,GND,TWR,APP,CTR,FSS |      |
| Administrator | Pilot,OBS,DEL,GND,TWR,APP,CTR,FSS |      |



### 特殊权限配置

你可以通过配置文件覆写管制权限与管制席位对照表  
注意!!! 这个字段会`覆盖`默认的对照表  
所以在明确的知道你在做什么之前, 不要修改这个配置  
配置文件字段为`rating`

```json5
{
  // 特殊权限配置
  "rating": {
    // 键为想要修改的权限识别名的权限值
    // 比如我想让Normal也可以上OBS席位, 也就是普通飞行员也可以以OBS身份连线
    // Normal的权限值是0, 那我的键就是0
    // 值为想要许可连线的席位的席位编码之和
    // 比如我想让飞行员可以正常连线, 也可以以OBS连线
    // 那么值就是 1 + 2 = 3
    // 如果我想他还能上个飞服(请勿模仿)
    // 那么值就是 1 + 2 + 128 = 131
    // 其他权限的对照表保持为默认
    // 你也可以将某个权限的值写为0来禁止使用该权限登录fsd
    "0": 3
  }
}
```



## 反馈办法

如您在使用FSD中遇到了任何/疑似bug的错误，请提交[issue](https://github.com/Flyleague-Collection/fsd-server/issues/new)在提交时，需要您按照以下步骤进行：

1. 对于可复现的bug:
   1. 在[配置文件](#配置文件简介)中打开`"debug_mode": true,`以启用log文件
   2. 重启FSD并复现此bug
   3. 将log文件上传至[issue](https://github.com/Flyleague-Collection/fsd-server/issues/new)
2. 对于不可复现的bug:
   1. 尽可能的用文字描述此bug并提交至[issue](https://github.com/Flyleague-Collection/fsd-server/issues/new)



## 链接

[HTTP API文档](https://3v26cojptv.apifox.cn/334394948e0)



## 开源协议

MIT License

Copyright (c) 2025 Half_nothing 

无附加条款。



## 行为准则

在[CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)中查阅