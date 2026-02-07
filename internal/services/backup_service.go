package services

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/database"
	"github.com/engigu/baihu-panel/internal/models"
)

type BackupService struct {
	settingsService *SettingsService
}

func NewBackupService() *BackupService {
	return &BackupService{
		settingsService: NewSettingsService(),
	}
}

const (
	BackupSection = "backup"
	BackupFileKey = "backup_file"
	BackupDir     = "./data/backups"
)

// tableConfig 表备份配置
type tableConfig struct {
	filename string
	export   func() (any, error)
	restore  func([]byte) error
}

func (s *BackupService) getTableConfigs() []tableConfig {
	return []tableConfig{
		{"tasks.json", s.exportTable(&[]models.Task{}, true), s.restoreTable(&[]models.Task{}, true)},
		{"task_logs.json", s.exportTable(&[]models.TaskLog{}, false), s.restoreTable(&[]models.TaskLog{}, false)},
		{"envs.json", s.exportTable(&[]models.EnvironmentVariable{}, true), s.restoreTable(&[]models.EnvironmentVariable{}, true)},
		{"scripts.json", s.exportTable(&[]models.Script{}, true), s.restoreTable(&[]models.Script{}, true)},
		{"settings.json", s.exportSettings, s.restoreSettings},
		{"send_stats.json", s.exportTable(&[]models.SendStats{}, false), s.restoreTable(&[]models.SendStats{}, false)},
		{"login_logs.json", s.exportTable(&[]models.LoginLog{}, false), s.restoreTable(&[]models.LoginLog{}, false)},
	}
}

func (s *BackupService) exportTable(dest any, unscoped bool) func() (any, error) {
	return func() (any, error) {
		db := database.DB
		if unscoped {
			db = db.Unscoped()
		}
		db.Find(dest)
		return dest, nil
	}
}

func (s *BackupService) restoreTable(dest any, unscoped bool) func([]byte) error {
	return func(data []byte) error {
		if err := json.Unmarshal(data, dest); err != nil {
			return err
		}
		return nil
	}
}

func (s *BackupService) exportSettings() (any, error) {
	var data []models.Setting
	database.DB.Where("section != ?", BackupSection).Find(&data)
	return data, nil
}

func (s *BackupService) restoreSettings(data []byte) error {
	var settings []models.Setting
	return json.Unmarshal(data, &settings)
}

// CreateBackup 创建备份
func (s *BackupService) CreateBackup() (string, error) {
	if err := os.MkdirAll(BackupDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("20060102_150405")
	zipPath := filepath.Join(BackupDir, fmt.Sprintf("backup_%s.zip", timestamp))

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 导出各表
	for _, cfg := range s.getTableConfigs() {
		data, err := cfg.export()
		if err != nil {
			return "", err
		}
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", err
		}
		w, err := zipWriter.Create(cfg.filename)
		if err != nil {
			return "", err
		}
		if _, err := w.Write(jsonData); err != nil {
			return "", err
		}
	}

	// 打包 scripts 文件夹
	scriptsDir := constant.ScriptsWorkDir
	if _, err := os.Stat(scriptsDir); err == nil {
		if err := s.addDirToZip(zipWriter, scriptsDir, "scripts"); err != nil {
			return "", err
		}
	}

	s.settingsService.Set(BackupSection, BackupFileKey, zipPath)
	return zipPath, nil
}

// Restore 恢复备份
func (s *BackupService) Restore(zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// 构建文件名到配置的映射
	configs := s.getTableConfigs()
	fileMap := make(map[string]*zip.File)
	for _, f := range r.File {
		fileMap[f.Name] = f
	}

	// 读取所有表数据
	tableData := make(map[string][]byte)
	for _, cfg := range configs {
		if f, ok := fileMap[cfg.filename]; ok {
			data, err := s.readZipFile(f)
			if err != nil {
				return err
			}
			tableData[cfg.filename] = data
		}
	}

	// 清空现有数据（物理删除）
	database.DB.Unscoped().Where("1=1").Delete(&models.Task{})
	database.DB.Unscoped().Where("1=1").Delete(&models.TaskLog{})
	database.DB.Unscoped().Where("1=1").Delete(&models.EnvironmentVariable{})
	database.DB.Unscoped().Where("1=1").Delete(&models.Script{})
	database.DB.Unscoped().Where("section != ?", BackupSection).Delete(&models.Setting{})
	database.DB.Unscoped().Where("1=1").Delete(&models.SendStats{})
	database.DB.Unscoped().Where("1=1").Delete(&models.LoginLog{})

	// 恢复数据
	s.restoreFromData(tableData, "tasks.json", &[]models.Task{})
	s.restoreFromData(tableData, "task_logs.json", &[]models.TaskLog{})
	s.restoreFromData(tableData, "envs.json", &[]models.EnvironmentVariable{})
	s.restoreFromData(tableData, "scripts.json", &[]models.Script{})
	s.restoreFromData(tableData, "settings.json", &[]models.Setting{})
	s.restoreFromData(tableData, "send_stats.json", &[]models.SendStats{})
	s.restoreFromData(tableData, "login_logs.json", &[]models.LoginLog{})

	// 恢复 scripts 文件夹
	s.restoreScriptsDir(r)

	return nil
}

func (s *BackupService) restoreFromData(tableData map[string][]byte, filename string, dest any) {
	if data, ok := tableData[filename]; ok {
		if err := json.Unmarshal(data, dest); err == nil {
			s.insertRecords(dest)
		}
	}
}

func (s *BackupService) insertRecords(records any) {
	switch v := records.(type) {
	case *[]models.Task:
		for _, r := range *v {
			database.DB.Create(&r)
		}
	case *[]models.TaskLog:
		for _, r := range *v {
			database.DB.Create(&r)
		}
	case *[]models.EnvironmentVariable:
		for _, r := range *v {
			database.DB.Create(&r)
		}
	case *[]models.Script:
		for _, r := range *v {
			database.DB.Create(&r)
		}
	case *[]models.Setting:
		for _, r := range *v {
			database.DB.Create(&r)
		}
	case *[]models.SendStats:
		for _, r := range *v {
			database.DB.Create(&r)
		}
	case *[]models.LoginLog:
		for _, r := range *v {
			database.DB.Create(&r)
		}
	}
}

func (s *BackupService) restoreScriptsDir(r *zip.ReadCloser) {
	scriptsDir := constant.ScriptsWorkDir
	for _, f := range r.File {
		if len(f.Name) > 8 && f.Name[:8] == "scripts/" {
			relPath := f.Name[8:]
			if relPath == "" {
				continue
			}
			fpath := filepath.Join(scriptsDir, relPath)
			if f.FileInfo().IsDir() {
				os.MkdirAll(fpath, 0755)
				continue
			}
			os.MkdirAll(filepath.Dir(fpath), 0755)
			if outFile, err := os.Create(fpath); err == nil {
				if rc, err := f.Open(); err == nil {
					io.Copy(outFile, rc)
					rc.Close()
				}
				outFile.Close()
			}
		}
	}
}

func (s *BackupService) readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func (s *BackupService) addDirToZip(zipWriter *zip.Writer, srcDir, prefix string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		zipPath := filepath.ToSlash(filepath.Join(prefix, relPath))
		if info.IsDir() {
			if relPath != "." {
				_, err := zipWriter.Create(zipPath + "/")
				return err
			}
			return nil
		}
		w, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(w, file)
		return err
	})
}

func (s *BackupService) GetBackupFile() string {
	var setting models.Setting
	if err := database.DB.Where("section = ? AND `key` = ?", BackupSection, BackupFileKey).First(&setting).Error; err != nil {
		return ""
	}
	return setting.Value
}

func (s *BackupService) ClearBackup() error {
	filePath := s.GetBackupFile()
	if filePath != "" {
		os.Remove(filePath)
		database.DB.Where("section = ? AND `key` = ?", BackupSection, BackupFileKey).Delete(&models.Setting{})
	}
	return nil
}
