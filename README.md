# DSH Tavern for fnOS

将 [dsh-tavern](https://github.com/flizzywine/dsh-tavern) 封装为飞牛 fnOS 原生应用包，并提供可视化管理界面。

本项目安装的是 **DSH Tavern**，不需要另外安装独立的 `deepseek.harness` 应用。Tavern 运行所需的 DSH Node.js 运行时会作为应用内部依赖部署到应用数据目录。

## 功能

- 服务启动、停止、重启、拉取更新与强制重建
- 进程组清理、异常退出巡检与端口占用检测
- Tavern Profile 与插件依赖自动配置
- 工作区列表、会话数量、更新时间及文件管理器定位
- 飞牛统一网关子路径访问
- 独立端口 HTTP/HTTPS 自适应反向代理
- 外部自定义访问地址与访问密码鉴权
- WebSocket 实时日志、语法高亮、自动滚动、清空与下载
- 日志轮转与应用设置热重载
- 服务端口、反向代理端口、网络代理及打开方式配置
- 支持插件安装、构建脚本放行、启用、禁用与卸载

## 安装

### 运行要求

- 飞牛 fnOS
- x86 设备使用 x86 FPK，ARM64 设备使用 ARM FPK
- 应用中心依赖 `nodejs_v24`

在 fnOS 应用中心选择“手动安装”，上传对应的 `.fpk` 文件即可。应用默认使用飞牛统一网关访问，也可以在应用设置中启用独立反向代理端口。

### FPK 文件

| 设备架构 | 文件 |
| --- | --- |
| x86-64 | `dsh.tavern-x86.fpk` 或 `dsh.tavern.fpk` |
| ARM64 | `dsh.tavern-arm.fpk` |

默认端口如下：

- DSH Tavern 内部服务：`3081`
- 独立反向代理：`3082`
- 飞牛统一网关：`/app/dsh-tavern/fngateway/`

端口和访问方式都可以在应用设置中修改。

## 更新与数据

安装更高版本的 `dsh.tavern.fpk` 可以直接更新应用。应用数据与配置保存在 fnOS 应用数据目录，更新包不会替换用户的 Tavern 工作区、会话和插件数据。

应用更新分为三个层级：

1. fnOS 应用包更新：安装更高版本的 `dsh.tavern.fpk`，更新管理器程序和内置资源。包内的 `tavern-source` 是 DSH Tavern 的离线运行资源，不会注册为独立的 `deepseek.harness` 应用。
2. DSH Tavern 程序更新：管理器中的“检查更新”检查 GitHub [dsh-tavern](https://github.com/flizzywine/dsh-tavern) 主分支提交，并负责停止、部署和重新启动服务。
3. 运行环境重建：“强制重建”重新安装 Tavern 依赖、Profile 和独立 DSH 运行时，不等同于更新 fnOS 应用包。

Tavern WebUI 内的上游更新提示只用于显示 DSH Tavern 提交；在 fnOS 中更新按钮由管理器统一执行，以避免与管理器托管的 3081 服务重复启动。它不会安装或更新独立的 `deepseek.harness` 应用。

## 构建

### Windows

构建环境需要 Node.js、Go、PowerShell 和 WSL Ubuntu：

```powershell
cd dsh-tavern.fnos
./build.cmd
```

### Linux

```bash
cd dsh-tavern.fnos
./build.sh
```

构建脚本会完成以下工作：

1. 构建 Vue 前端资源。
2. 交叉编译 Linux amd64 与 arm64 后端。
3. 将内置 Tavern 源码包和 fnOS 生命周期脚本打入 FPK。
4. 生成 `dsh.tavern-x86.fpk` 与 `dsh.tavern-arm.fpk`。

## 项目结构

```text
dsh-tavern.fnos/
├── api.go                         # REST API 与 WebSocket 通道
├── harness.go                     # 服务生命周期与启动状态机
├── process.go                     # 进程组控制与进程清理
├── plugins.go                     # 插件安装、启停与卸载
├── profile.go                     # Tavern Profile 配置与补丁
├── workspace.go                   # 工作区数据读取
├── proxy.go                       # 独立反向代理与访问鉴权
├── fngateway.go                   # fnOS 统一网关子路径适配
├── frontend/                      # Vue 3 管理界面
├── fnpack/                        # fnOS 应用包清单与生命周期脚本
├── tavern-source/                 # 内置 dsh-tavern 离线运行资源（打入 FPK）
└── REVERSE_PROXY_ADAPTATION.md    # 反向代理适配说明
```

## 反向代理说明

DSH 的客户端插件资源使用 `/plugins/??...&rev=...` 组合资源协议。项目的两个反代入口都会保留该原始查询串，并处理 fnOS 网关子路径下的 HTML、静态资源、Fetch、XHR、WebSocket 和 SSE 请求。

详细实现请参阅 [REVERSE_PROXY_ADAPTATION.md](./REVERSE_PROXY_ADAPTATION.md)。

## 感谢

- [deepseek.harness.fnos](https://github.com/yuexps/deepseek.harness.fnos)：fnOS 原生应用包结构与实现参考。
- [dsh-tavern](https://github.com/flizzywine/dsh-tavern)：Tavern 主项目及其功能实现。

## 相关项目

- [dsh-tavern](https://github.com/flizzywine/dsh-tavern)
- [deepseek.harness.fnos](https://github.com/yuexps/deepseek.harness.fnos)
