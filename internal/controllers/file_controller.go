package controllers

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/engigu/baihu-panel/internal/utils"

	"github.com/gin-gonic/gin"
)

var (
	extractZip   = utils.ExtractZip
	extractTar   = utils.ExtractTar
	extractTarGz = utils.ExtractTarGz
)

type FileController struct {
	workDir string
}

func NewFileController(workDir string) *FileController {
	os.MkdirAll(workDir, 0755)
	absPath, err := filepath.Abs(workDir)
	if err != nil {
		absPath = workDir
	}
	return &FileController{workDir: absPath}
}

type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"`
}

func (fc *FileController) GetFileTree(c *gin.Context) {
	root := &FileNode{
		Name:     filepath.Base(fc.workDir),
		Path:     "",
		IsDir:    true,
		Children: []*FileNode{},
	}

	err := filepath.WalkDir(fc.workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == fc.workDir {
			return nil
		}

		// 过滤 __pycache__ 文件夹
		if d.IsDir() && d.Name() == "__pycache__" {
			return filepath.SkipDir
		}

		relPath, _ := filepath.Rel(fc.workDir, path)
		parts := strings.Split(relPath, string(filepath.Separator))

		current := root
		for i, part := range parts {
			found := false
			for _, child := range current.Children {
				if child.Name == part {
					current = child
					found = true
					break
				}
			}
			if !found {
				isLast := i == len(parts)-1
				isDir := !isLast || d.IsDir()
				node := &FileNode{
					Name:  part,
					Path:  strings.Join(parts[:i+1], "/"),
					IsDir: isDir,
				}
				if isDir {
					node.Children = []*FileNode{}
				}
				current.Children = append(current.Children, node)
				current = node
			}
		}
		return nil
	})

	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, root.Children)
}

func (fc *FileController) GetFileContent(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		utils.BadRequest(c, "path参数必填")
		return
	}

	fullPath := filepath.Join(fc.workDir, filepath.Clean(filePath))
	if !strings.HasPrefix(fullPath, fc.workDir) {
		utils.Forbidden(c, "访问被拒绝")
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		utils.NotFound(c, "文件不存在")
		return
	}

	utils.Success(c, gin.H{
		"path":    filePath,
		"content": string(content),
	})
}

func (fc *FileController) SaveFileContent(c *gin.Context) {
	var req struct {
		Path    string `json:"path" binding:"required"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	fullPath := filepath.Join(fc.workDir, filepath.Clean(req.Path))
	if !strings.HasPrefix(fullPath, fc.workDir) {
		utils.Forbidden(c, "访问被拒绝")
		return
	}

	os.MkdirAll(filepath.Dir(fullPath), 0755)

	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.SuccessMsg(c, "保存成功")
}

func (fc *FileController) CreateFile(c *gin.Context) {
	var req struct {
		Path  string `json:"path" binding:"required"`
		IsDir bool   `json:"isDir"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	fullPath := filepath.Join(fc.workDir, filepath.Clean(req.Path))
	if !strings.HasPrefix(fullPath, fc.workDir) {
		utils.Forbidden(c, "访问被拒绝")
		return
	}

	if req.IsDir {
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			utils.ServerError(c, err.Error())
			return
		}
	} else {
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err := os.WriteFile(fullPath, []byte(""), 0644); err != nil {
			utils.ServerError(c, err.Error())
			return
		}
	}

	utils.SuccessMsg(c, "创建成功")
}

func (fc *FileController) DeleteFile(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	fullPath := filepath.Join(fc.workDir, filepath.Clean(req.Path))
	if !strings.HasPrefix(fullPath, fc.workDir) {
		utils.Forbidden(c, "访问被拒绝")
		return
	}

	if err := os.RemoveAll(fullPath); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.SuccessMsg(c, "删除成功")
}

func (fc *FileController) RenameFile(c *gin.Context) {
	var req struct {
		OldPath string `json:"oldPath" binding:"required"`
		NewPath string `json:"newPath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	oldFull := filepath.Join(fc.workDir, filepath.Clean(req.OldPath))
	newFull := filepath.Join(fc.workDir, filepath.Clean(req.NewPath))

	if !strings.HasPrefix(oldFull, fc.workDir) || !strings.HasPrefix(newFull, fc.workDir) {
		utils.Forbidden(c, "访问被拒绝")
		return
	}

	// 确保目标目录存在
	os.MkdirAll(filepath.Dir(newFull), 0755)

	if err := os.Rename(oldFull, newFull); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.SuccessMsg(c, "移动成功")
}

// UploadArchive handles archive file upload and extraction
func (fc *FileController) UploadArchive(c *gin.Context) {
	targetDir := c.PostForm("path")

	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".zip" && ext != ".tar" && ext != ".gz" && ext != ".tgz" {
		utils.BadRequest(c, "仅支持 zip、tar、gz、tgz 格式")
		return
	}

	// 确定解压目标目录
	extractDir := fc.workDir
	if targetDir != "" {
		extractDir = filepath.Join(fc.workDir, filepath.Clean(targetDir))
		if !strings.HasPrefix(extractDir, fc.workDir) {
			utils.Forbidden(c, "访问被拒绝")
			return
		}
	}
	os.MkdirAll(extractDir, 0755)

	// 保存临时文件
	tempFile := filepath.Join(os.TempDir(), file.Filename)
	if err := c.SaveUploadedFile(file, tempFile); err != nil {
		utils.ServerError(c, "保存文件失败")
		return
	}
	defer os.Remove(tempFile)

	// 解压文件
	var extractErr error
	switch {
	case ext == ".zip":
		extractErr = extractZip(tempFile, extractDir)
	case ext == ".tar":
		extractErr = extractTar(tempFile, extractDir)
	case ext == ".gz" || ext == ".tgz":
		extractErr = extractTarGz(tempFile, extractDir)
	}

	if extractErr != nil {
		utils.ServerError(c, "解压失败: "+extractErr.Error())
		return
	}

	utils.SuccessMsg(c, "导入成功")
}

// UploadFiles handles multiple file uploads
func (fc *FileController) UploadFiles(c *gin.Context) {
	targetDir := c.PostForm("path")

	// 确定目标目录
	destDir := fc.workDir
	if targetDir != "" {
		destDir = filepath.Join(fc.workDir, filepath.Clean(targetDir))
		if !strings.HasPrefix(destDir, fc.workDir) {
			utils.Forbidden(c, "访问被拒绝")
			return
		}
	}
	os.MkdirAll(destDir, 0755)

	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}

	files := form.File["files"]
	paths := form.Value["paths"] // 相对路径数组，用于保持文件夹结构

	if len(files) == 0 {
		utils.BadRequest(c, "请选择文件")
		return
	}

	for i, file := range files {
		// 获取相对路径（如果有）
		relPath := file.Filename
		if i < len(paths) && paths[i] != "" {
			relPath = paths[i]
		}

		// 构建完整路径
		fullPath := filepath.Join(destDir, filepath.Clean(relPath))

		// 安全检查
		if !strings.HasPrefix(fullPath, fc.workDir) {
			continue
		}

		// 确保父目录存在
		os.MkdirAll(filepath.Dir(fullPath), 0755)

		// 保存文件
		if err := c.SaveUploadedFile(file, fullPath); err != nil {
			utils.ServerError(c, "保存文件失败: "+err.Error())
			return
		}
	}

	utils.SuccessMsg(c, "上传成功")
}
