/*
 * Copyright 2021 Huawei Technologies Co., Ltd.
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
	"fileSystem/models"
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

func (c *BaseController) insertOrUpdateCheckRecord(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int, checkStatusResponse CheckStatusResponse) error {
	fileRecord := &models.ImageDB{
		ImageId:        imageId,
		FileName:       fileName,
		UserId:         userId,
		StorageMedium:  storageMedium,
		SaveFileName:   saveFileName,
		SlimStatus:     slimStatus,
		Checksum:       checkStatusResponse.CheckInformation.Checksum,
		CheckResult:    checkStatusResponse.CheckInformation.CheckResult,
		CheckMsg:       checkStatusResponse.Msg,
		CheckStatus:    checkStatusResponse.Status,
		ImageEndOffset: checkStatusResponse.CheckInformation.ImageInformation.ImageEndOffset,
		CheckErrors:    checkStatusResponse.CheckInformation.ImageInformation.CheckErrors,
		Format:         checkStatusResponse.CheckInformation.ImageInformation.Format,
	}
	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")
	if err != nil && err.Error() != util.LastInsertIdNotSupported {
		log.Error(util.FailToRecordToDB)
		return err
	}
	log.Info(util.FileRecord, fileRecord)
	return nil
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