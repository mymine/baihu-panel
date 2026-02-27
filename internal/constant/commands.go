package constant

// CommandInfo 定义了终端可用命令的说明信息
type CommandInfo struct {
	Name        string
	Description string
}

// Commands 是系统的可用业务命令说明列表
var Commands = []CommandInfo{
	// {
	// 	Name:        "server",
	// 	Description: "启动后台服务进程",
	// },
	{
		Name:        "reposync",
		Description: "同步远程 Git 仓库或文件到本地",
	},
	{
		Name:        "resetpwd",
		Description: "重置 admin 用户密码（需要二次确认）",
	},
	{
		Name:        "restore",
		Description: "从本地 zip 文件中全量恢复系统级备份数据",
	},
}
