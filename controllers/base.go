package controllers

import (
	"fileSystem/pkg/dbAdpater"
	"fileSystem/util"
	"github.com/astaxie/beego"
	log "github.com/sirupsen/logrus"
)


type BaseController struct {
	beego.Controller
	Db dbAdpater.Database
}

// To display log for received message
func (c *BaseController) displayReceivedMsg(clientIp string) {
	log.Info("Received message from ClientIP [" + clientIp + util.Operation + c.Ctx.Request.Method + "]" +
		util.Resource + c.Ctx.Input.URL() + "]")
}


// Write response
func (c *BaseController) writeResponse(msg string, code int) {
//	c.Data["json"] = msg
	c.Ctx.ResponseWriter.WriteHeader(code)

//	c.ServeJSON()
}

// Write error response
func (c *BaseController) writeErrorResponse(errMsg string, code int) {
	log.Error(errMsg)
	c.writeResponse(errMsg, code)
}

// Handled logging for error case
func (c *BaseController) HandleLoggingForError(clientIp string, code int, errMsg string) {
	c.writeErrorResponse(errMsg, code)
	log.Info("Response message for ClientIP [" + clientIp + util.Operation + c.Ctx.Request.Method + "]" +
		util.Resource + c.Ctx.Input.URL() + "] Result [Failure: " + errMsg + ".]")
}