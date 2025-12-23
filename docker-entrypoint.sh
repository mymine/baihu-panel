#!/bin/sh
set -e

PYTHON_VENV_DIR="/app/envs/python"
NODE_ENV_DIR="/app/envs/node"

# ============================
# 创建基础目录
# ============================
mkdir -p /app/data /app/data/scripts /app/configs /app/envs

# ============================
# Python 虚拟环境
# ============================
if [ ! -x "$PYTHON_VENV_DIR/bin/python" ]; then
    echo "[entrypoint] Creating Python venv..."
    python3 -m venv "$PYTHON_VENV_DIR"
    "$PYTHON_VENV_DIR/bin/pip" install --upgrade pip setuptools wheel
    "$PYTHON_VENV_DIR/bin/pip" config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
else
    echo "[entrypoint] Python venv exists"
fi

# ============================
# Node 本地环境（npm prefix 方案）
# ============================
if [ ! -d "$NODE_ENV_DIR/node_modules" ]; then
    echo "[entrypoint] Initializing Node environment..."
    mkdir -p "$NODE_ENV_DIR"
    cd "$NODE_ENV_DIR"
    npm init -y >/dev/null 2>&1
else
    echo "[entrypoint] Node environment exists"
fi

# ============================
# npm / Node 内存与行为控制（关键）
# ============================
export NODE_OPTIONS="--max-old-space-size=256"

npm config set registry https://registry.npmmirror.com
npm config set audit false
npm config set fund false
npm config set progress false
npm config set maxsockets 2

# ============================
# 环境变量注入（等价于 activate）
# ============================
export PATH="$PYTHON_VENV_DIR/bin:$NODE_ENV_DIR/node_modules/.bin:$PATH"
export NODE_PATH="$NODE_ENV_DIR/node_modules"

# ============================
# 打印确认（非常建议保留）
# ============================
echo "[entrypoint] python: $(which python)"
echo "[entrypoint] pip: $(which pip)"
echo "[entrypoint] node: $(which node)"
echo "[entrypoint] npm root: $(npm root --prefix $NODE_ENV_DIR)"

# ============================
# 启动应用
# ============================
exec /app/baihu