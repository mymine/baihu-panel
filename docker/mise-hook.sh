#!/bin/bash

# 面向终端的交互式 Hook 拦截器
# 目的是为了在终端手动执行 `mise exec node -- node xxx` 时也能具有与 Go 系统底层相同的 NODE_PATH 环境
mise() {
  if [[ "$1" == "exec" || "$1" == "x" ]]; then
    local is_node=false
    for arg in "$@"; do
      if [[ "$arg" == *"node"* ]]; then
        is_node=true
        break
      fi
    done

    if [[ "$is_node" == true ]]; then
      local node_bin=$(command mise which node 2>/dev/null)
      if [[ -n "$node_bin" && "$node_bin" == *"/bin/node" ]]; then
        # 截取路径并附加到当前进程执行环境的前置变量
        NODE_PATH="${node_bin%/bin/node}/lib/node_modules" command mise "$@"
        return $?
      fi
    fi
  fi
  # 非 Node 场景或拦截失败，正常向下执行
  command mise "$@"
}

# 导出此函数，供所有子环境继承使用
export -f mise
