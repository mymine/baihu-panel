package controllers

import (
	"strconv"

	"github.com/engigu/baihu-panel/internal/database"
	"github.com/engigu/baihu-panel/internal/models"
	"github.com/engigu/baihu-panel/internal/utils"

	"github.com/gin-gonic/gin"
)

type LogController struct{}

func NewLogController() *LogController {
	return &LogController{}
}

type TaskLogResponse struct {
	ID        uint              `json:"id"`
	TaskID    uint              `json:"task_id"`
	TaskName  string            `json:"task_name"`
	TaskType  string            `json:"task_type"`
	Command   string            `json:"command"`
	Status    string            `json:"status"`
	Duration  int64             `json:"duration"`
	StartTime *models.LocalTime `json:"start_time"`
	EndTime   *models.LocalTime `json:"end_time"`
	CreatedAt models.LocalTime  `json:"created_at"`
}

func (lc *LogController) GetLogs(c *gin.Context) {
	p := utils.ParsePagination(c)
	taskID, _ := strconv.Atoi(c.DefaultQuery("task_id", "0"))
	taskName := c.DefaultQuery("task_name", "")

	var logs []models.TaskLog
	var total int64

	query := database.DB.Model(&models.TaskLog{})
	if taskID > 0 {
		query = query.Where("task_id = ?", taskID)
	}

	// 按任务名称过滤
	if taskName != "" {
		var taskIDs []uint
		database.DB.Model(&models.Task{}).Where("name LIKE ?", "%"+taskName+"%").Pluck("id", &taskIDs)
		if len(taskIDs) > 0 {
			query = query.Where("task_id IN ?", taskIDs)
		} else {
			utils.PaginatedResponse(c, []TaskLogResponse{}, 0, p)
			return
		}
	}

	query.Count(&total)
	query.Order("id DESC").Offset(p.Offset()).Limit(p.PageSize).Find(&logs)

	taskIDList := make([]uint, 0)
	for _, log := range logs {
		taskIDList = append(taskIDList, log.TaskID)
	}

	var tasks []models.Task
	database.DB.Where("id IN ?", taskIDList).Find(&tasks)
	taskMap := make(map[uint]models.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	result := make([]TaskLogResponse, len(logs))
	for i, log := range logs {
		task := taskMap[log.TaskID]
		taskType := task.Type
		if taskType == "" {
			taskType = "task"
		}
		result[i] = TaskLogResponse{
			ID:        log.ID,
			TaskID:    log.TaskID,
			TaskName:  task.Name,
			TaskType:  taskType,
			Command:   log.Command,
			Status:    log.Status,
			Duration:  log.Duration,
			StartTime: log.StartTime,
			EndTime:   log.EndTime,
			CreatedAt: log.CreatedAt,
		}
	}

	utils.PaginatedResponse(c, result, total, p)
}

func (lc *LogController) GetLogDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的日志ID")
		return
	}

	var log models.TaskLog
	if err := database.DB.First(&log, id).Error; err != nil {
		utils.NotFound(c, "日志不存在")
		return
	}

	utils.Success(c, gin.H{
		"id":         log.ID,
		"task_id":    log.TaskID,
		"command":    log.Command,
		"output":     log.Output,
		"status":     log.Status,
		"duration":   log.Duration,
		"start_time": log.StartTime,
		"end_time":   log.EndTime,
		"created_at": log.CreatedAt,
	})
}
