/*
 * Copyright 202 Huawei Technologies Co., Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// @Title  controllers
// @Description  base controller for filesystem
// @Author  GuoZhen Gao (2021/6/30 10:40)
package controllers

import (
	"fileSystem/pkg/dbAdpater"
	"fileSystem/util"
	"github.com/astaxie/beego"
	log "github.com/sirupsen/logrus"
)

// BaseController   Define the base for other controllers
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
	c.Data["json"] = msg
	c.Ctx.ResponseWriter.WriteHeader(code)
	c.ServeJSON()
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

// Handled logging for success case
func (c *BaseController) handleLoggingForSuccess(clientIp string, msg string) {
	log.Info("Response message for ClientIP [" + clientIp + util.Operation + c.Ctx.Request.Method + "]" +
		util.Resource + c.Ctx.Input.URL() + "] Result [Success: " + msg + ".]")
}