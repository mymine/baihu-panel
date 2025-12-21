package controllers

import (
	"strconv"

	"baihu/internal/services"
	"baihu/internal/utils"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	taskService *services.TaskService
	cronService *services.CronService
}

func NewTaskController(taskService *services.TaskService, cronService *services.CronService) *TaskController {
	return &TaskController{
		taskService: taskService,
		cronService: cronService,
	}
}

func (tc *TaskController) CreateTask(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Command     string `json:"command" binding:"required"`
		Schedule    string `json:"schedule" binding:"required"`
		Timeout     int    `json:"timeout"`
		CleanConfig string `json:"clean_config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := tc.cronService.ValidateCron(req.Schedule); err != nil {
		utils.BadRequest(c, "无效的cron表达式: "+err.Error())
		return
	}

	task := tc.taskService.CreateTask(req.Name, req.Command, req.Schedule, req.Timeout, req.CleanConfig)
	tc.cronService.AddTask(task)

	utils.Success(c, task)
}

func (tc *TaskController) GetTasks(c *gin.Context) {
	p := utils.ParsePagination(c)
	name := c.DefaultQuery("name", "")

	tasks, total := tc.taskService.GetTasksWithPagination(p.Page, p.PageSize, name)
	utils.PaginatedResponse(c, tasks, total, p)
}

func (tc *TaskController) GetTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	task := tc.taskService.GetTaskByID(id)
	if task == nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	utils.Success(c, task)
}

func (tc *TaskController) UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Command     string `json:"command"`
		Schedule    string `json:"schedule"`
		Timeout     int    `json:"timeout"`
		CleanConfig string `json:"clean_config"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if req.Schedule != "" {
		if err := tc.cronService.ValidateCron(req.Schedule); err != nil {
			utils.BadRequest(c, "无效的cron表达式: "+err.Error())
			return
		}
	}

	task := tc.taskService.UpdateTask(id, req.Name, req.Command, req.Schedule, req.Timeout, req.CleanConfig, req.Enabled)
	if task == nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	if task.Enabled {
		tc.cronService.AddTask(task)
	} else {
		tc.cronService.RemoveTask(task.ID)
	}

	utils.Success(c, task)
}

func (tc *TaskController) DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	tc.cronService.RemoveTask(uint(id))

	success := tc.taskService.DeleteTask(id)
	if !success {
		utils.NotFound(c, "任务不存在")
		return
	}

	utils.SuccessMsg(c, "删除成功")
}
