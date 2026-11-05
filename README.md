# macOS Auto Darkmode Switcher

一个基于 Go 语言开发的轻量级 macOS 后台守护程序。它通过实时轮询当前屏幕亮度，当亮度低于设定阈值时自动将系统切换为**暗黑模式（Dark Mode）**，超过阈值时切回**浅色模式（Light Mode）**。

## 核心特性

* **标准工程架构**：遵循 Go 官方 Standard Go Project Layout（`cmd/`、`pkg/`、`deploy/`）。
* **精准解析**：严格适配 `brightness` 命令行工具的标准输出，数值解析安全可靠。
* **系统级常驻**：通过 macOS 原生的 `launchd` 实现后台守护，开机自启、崩溃自动拉起。
* **无缝部署**：提供完善的 `Makefile`，一键完成编译、路径注入、服务注册与启动。

---

## 前置依赖

在编译和安装本项目之前，请确保系统中已安装 `brightness` 命令行工具：

```bash
brew install brightness
# or build from source
git clone https://github.com/nriley/brightness.git
cd brightness
make
sudo make install
```

---

## 项目目录结构

```text
auto-darkmode-switcher/
├── go.mod
├── Makefile                                # 工程生命周期管理
├── README.md                               # 项目说明文档
├── deploy/
│   └── com.proteu.autodarkmodeswitcher.plist # launchd 守护进程模板
├── cmd/
│   └── auto-darkmode-switcher/
│       └── main.go                         # 程序入口与轮询主循环
└── pkg/
    └── brightness/
        ├── brightness.go                   # 亮度采集与解析
        └── darkmode.go                     # 系统外观状态查询与切换
```

---

## 快速开始与部署

本项目完全由 `Makefile` 驱动，支持一键式生命周期管理：

### 1. 编译程序
仅在本地编译二进制文件（输出至 `bin/auto-darkmode-switcher`）：
```bash
make build
```

### 2. 安装并启动后台服务
一键完成以下动作：
* 创建用户局部二进制目录 (`~/.local/bin`) 与日志目录 (`~/.local/var/log`)。
* 拷贝编译后的二进制文件。
* 动态替换 `deploy/` 下 `plist` 模板中的绝对路径，并注册到 `~/Library/LaunchAgents/`。
* 启动 `launchd` 守护进程。
```bash
make install
```

### 3. 查看实时运行日志
```bash
make logs
```

### 4. 卸载服务
停止并注销后台守护进程，清理相关二进制文件：
```bash
make uninstall
```

---

## 参数配置

默认情况下，守护进程以 **`0.7` 亮度阈值**、**`0.035秒` 偏移量**、**`5秒` 检测间隔**、**`3秒` 防抖间隔** 运行。

如需调整参数，可以修改 `deploy/com.proteu.autodarkmodeswitcher.plist` 文件中的 `ProgramArguments` 数组：
* `-v`: 亮度切换阈值（范围 `0.0 ~ 1.0`）
* `-o`: 亮度切换偏移量（范围 `0.0 ~ 0.1`）
* `-i`: 轮询检测间隔时间（单位：秒）
* `-d`: 防抖间隔时间（单位：秒）

修改后重新执行 `make install` 即可生效。