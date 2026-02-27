package restore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/engigu/baihu-panel/internal/bootstrap"
	"github.com/engigu/baihu-panel/internal/services"
)

func Run(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: baihu restore <backup_file.zip>")
		os.Exit(1)
	}

	backupFile := args[0]
	absPath, err := filepath.Abs(backupFile)
	if err != nil {
		fmt.Printf("文件路径解析失败: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("错误: 备份文件 '%s' 不存在\n", absPath)
		os.Exit(1)
	}

	// 必须初始化环境与数据库才能恢复数据
	bootstrap.InitBasic()

	backupService := services.NewBackupService()
	fmt.Printf("正在从 '%s' 恢复系统数据，请勿强制中断...\n", absPath)
	err = backupService.Restore(absPath)
	if err != nil {
		fmt.Printf("恢复备份失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("系统备份恢复成功！")
	fmt.Println("注意：部分设定可能需要重启后台服务后才能完全生效。")
	fmt.Println("--------------------------------------------------")
}
