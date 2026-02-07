#!/bin/sh
set -e

PYTHON_VENV_DIR="/app/envs/python"
NODE_ENV_DIR="/app/envs/node"

echo "[entrypoint] Starting environment initialization..."

# ============================
# 创建基础目录
# ============================
mkdir -p \
  /app/data \
  /app/data/scripts \
  /app/configs \
  /app/envs \
  "$NODE_ENV_DIR"

# ============================
# Python 虚拟环境（后台并行）
# ============================
(
    if [ ! -x "$PYTHON_VENV_DIR/bin/python" ]; then
        echo "[entrypoint][Python3] Creating Python virtual environment..."
        python3 -m venv "$PYTHON_VENV_DIR"
        "$PYTHON_VENV_DIR/bin/pip" config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
        "$PYTHON_VENV_DIR/bin/pip" install --upgrade pip setuptools wheel
        echo "[entrypoint][Python3] Python venv created"
    else
        echo "[entrypoint][Python3] Python venv exists"
    fi
) &
PYTHON_PID=$!

# ============================
# Node 环境（后台并行）
# ============================
(
    echo "[entrypoint][Nodejs] Initializing Node npm prefix..."
    # npm 全局安装目录 → /app/envs/node
    npm config set prefix "$NODE_ENV_DIR"
    # npm 行为优化（防内存暴涨）
    npm config set registry https://registry.npmmirror.com
    npm config set audit false
    npm config set fund false
    npm config set progress false
    npm config set maxsockets 2
    echo "[entrypoint][Nodejs] Node npm configured"
) &
NODE_PID=$!

# 等待两个后台任务完成
echo "[entrypoint] Waiting for environment initialization..."
wait $PYTHON_PID
wait $NODE_PID
echo "[entrypoint] Environment initialization completed"

# Node 内存限制
export NODE_OPTIONS="--max-old-space-size=256"

# ============================
# 环境变量注入（等价 activate）
# ============================
export PATH="$PYTHON_VENV_DIR/bin:$NODE_ENV_DIR/bin:$PATH"
export NODE_PATH="$NODE_ENV_DIR/lib/node_modules"
export PYTHONPATH=/app/data/scripts:$PYTHONPATH

# ============================
# 打印确认
# ============================
echo "[entrypoint][Python3] python: $(which python)"
echo "[entrypoint][Python3] pip: $(which pip)"
echo "[entrypoint][Nodejs] node: $(which node)"
echo "[entrypoint][Nodejs] npm prefix: $(npm config get prefix)"
echo "[entrypoint][Nodejs] npm root: $(npm root -g)"

# ============================
# 启动应用
# ============================
cd /app
exec ./baihu