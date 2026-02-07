package controllers

import (
	"strconv"

	"github.com/engigu/baihu-panel/internal/services"
	"github.com/engigu/baihu-panel/internal/utils"

	"github.com/gin-gonic/gin"
)

type ScriptController struct {
	scriptService *services.ScriptService
}

func NewScriptController(scriptService *services.ScriptService) *ScriptController {
	return &ScriptController{scriptService: scriptService}
}

func (sc *ScriptController) CreateScript(c *gin.Context) {
	userID := 1

	var req struct {
		Name    string `json:"name" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	script := sc.scriptService.CreateScript(req.Name, req.Content, userID)
	utils.Success(c, script)
}

func (sc *ScriptController) GetScripts(c *gin.Context) {
	userID := 1
	scripts := sc.scriptService.GetScriptsByUserID(userID)
	utils.Success(c, scripts)
}

func (sc *ScriptController) GetScript(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的脚本ID")
		return
	}

	script := sc.scriptService.GetScriptByID(id)
	if script == nil {
		utils.NotFound(c, "脚本不存在")
		return
	}

	utils.Success(c, script)
}

func (sc *ScriptController) UpdateScript(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的脚本ID")
		return
	}

	var req struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	script := sc.scriptService.UpdateScript(id, req.Name, req.Content)
	if script == nil {
		utils.NotFound(c, "脚本不存在")
		return
	}

	utils.Success(c, script)
}

func (sc *ScriptController) DeleteScript(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的脚本ID")
		return
	}

	success := sc.scriptService.DeleteScript(id)
	if !success {
		utils.NotFound(c, "脚本不存在")
		return
	}

	utils.SuccessMsg(c, "删除成功")
}
