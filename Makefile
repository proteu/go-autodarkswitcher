APP_NAME = auto-darkmode-switcher
BIN_DIR = $(HOME)/.local/bin
LOG_DIR = $(HOME)/.local/var/log
PLIST_SRC = deploy/com.proteu.autodarkmodeswitcher.plist
PLIST_DEST = $(HOME)/Library/LaunchAgents/com.proteu.autodarkmodeswitcher.plist

.PHONY: all build install uninstall logs clean

all: build

build:
	@echo "==> 正在编译 $(APP_NAME)..."
	@mkdir -p bin
	go build -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

install: build
	@echo "==> 正在部署系统守护进程..."
	@mkdir -p $(BIN_DIR)
	@mkdir -p $(LOG_DIR)
	@mkdir -p $(HOME)/Library/LaunchAgents
	
	# 1. 拷贝二进制文件到用户局部环境
	cp bin/$(APP_NAME) $(BIN_DIR)/$(APP_NAME)
	
	# 2. 如果服务已存在，先卸载
	@launchctl unload $(PLIST_DEST) 2>/dev/null || true
	
	# 3. 动态替换 plist 中的绝对路径并写入目标目录
	@sed -e 's|__BIN_PATH__|$(BIN_DIR)/$(APP_NAME)|g' \
	     -e 's|__LOG_PATH__|$(LOG_DIR)|g' \
	     $(PLIST_SRC) > $(PLIST_DEST)
	
	# 4. 加载并启动服务
	launchctl load $(PLIST_DEST)
	@echo "==> 部署成功！程序已在后台运行。"

uninstall:
	@echo "==> 正在卸载守护进程..."
	@launchctl unload $(PLIST_DEST) 2>/dev/null || true
	@rm -f $(PLIST_DEST)
	@rm -f $(BIN_DIR)/$(APP_NAME)
	@echo "==> 卸载完成。"

logs:
	@echo "==> 实时查看运行日志 (Ctrl+C 退出)..."
	tail -f $(LOG_DIR)/$(APP_NAME).log

clean:
	@echo "==> 清理编译缓存..."
	rm -rf bin/

