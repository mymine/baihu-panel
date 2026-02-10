// 应用路径常量
export const PATHS = {
  // 脚本文件目录
  SCRIPTS_DIR: '/app/data/scripts',
  // 数据目录
  DATA_DIR: '/app/data',
  // 配置目录
  CONFIGS_DIR: '/app/configs',
  // 环境目录
  ENVS_DIR: '/app/envs',
} as const

// 文件扩展名对应的运行命令
export const FILE_RUNNERS: Record<string, string> = {
  py: 'python',
  js: 'node',
  sh: 'bash',
  bash: 'bash',
} as const

// 任务状态
export const TASK_STATUS = {
  SUCCESS: 'success',
  FAILED: 'failed',
  RUNNING: 'running',
  PENDING: 'pending',
  TIMEOUT: 'timeout',
  CANCELLED: 'cancelled',
} as const

// 任务类型
export const TASK_TYPE = {
  NORMAL: 'task',
  REPO: 'repo',
} as const

// Agent 状态
export const AGENT_STATUS = {
  ONLINE: 'online',
  OFFLINE: 'offline',
} as const
