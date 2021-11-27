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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fileSystem/models"
	"fileSystem/pkg/dbAdpater"
	"fileSystem/util"
	"github.com/astaxie/beego"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
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

func (c *BaseController) CronGetCheck(requestIdCheck string, imageId string, originalName string, userId string, storageMedium string, saveFileName string) {
	log.Warn("go routine is here")
	//此时瘦身结束，查看Check Response详情
	isCheckFinished := false
	checkTimes := 120
	for !isCheckFinished && checkTimes > 0 {
		checkTimes--
		if len(requestIdCheck) == 0 {
			log.Error("after POST check to imageOps, check requestId is still empty")
			c.writeErrorResponse("after POST check to imageOps, check requestId is still empty", util.StatusInternalServerError)
			return
		}

		checkStatusResponse, err := c.GetToCheck(requestIdCheck)
		if err != nil {
			log.Error("Fail to request to imageOps GET with requestId: " + requestIdCheck)
			c.writeErrorResponse("Fail to request to imageOps GET with requestId: "+requestIdCheck, util.StatusInternalServerError)
			return
		}

		if checkStatusResponse.Status == util.CheckInProgress { // check in progress
			time.Sleep(time.Duration(30) * time.Second)
			continue
		} else {
			isCheckFinished = true
			var imageFileDb models.ImageDB
			log.Info("query db ok.")
			_, err := c.Db.QueryTable("image_d_b", &imageFileDb, "image_id__exact", imageId)
			if err != nil {
				c.writeErrorResponse("fail to query database", util.StatusNotFound)
				return
			}
			slimStatus := imageFileDb.SlimStatus
			err = c.insertOrUpdateCheckRecord(imageId, originalName, userId, storageMedium, saveFileName, slimStatus, checkStatusResponse)
			if err != nil {
				log.Error(util.FailedToInsertDataToDB)
				c.writeErrorResponse(util.FailToInsertRequestCheck, util.StatusInternalServerError)
				return
			}
		}
	}
}

func (c *BaseController) GetToCheck(requestIdCheck string) (CheckStatusResponse, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	responseCheck, err := client.Get("http://localhost:5000/api/v1/vmimage/check/" + requestIdCheck)
	if err != nil {
		log.Error("fail to request imageOps check")
		c.writeErrorResponse("fail to request imageOps check", util.StatusInternalServerError)
		return CheckStatusResponse{}, err
	}
	defer responseCheck.Body.Close()
	bodyCheck, err := ioutil.ReadAll(responseCheck.Body)
	var checkStatusResponse CheckStatusResponse
	err = json.Unmarshal(bodyCheck, &checkStatusResponse)
	if err != nil {
		log.Error("GET to image check failed to unmarshal request")
		c.writeErrorResponse("GET to image check failed to unmarshal request", util.BadRequest)
		return CheckStatusResponse{}, err
	}
	return checkStatusResponse, nil
}

func (c *BaseController) PostToCheck(saveFileName string) (CheckResponse, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	var formConfigMap map[string]string
	formConfigMap = make(map[string]string)
	formConfigMap["inputImageName"] = saveFileName
	requestJson, _ := json.Marshal(formConfigMap)
	requestBody := bytes.NewReader(requestJson)
	response, err := client.Post("http://localhost:5000/api/v1/vmimage/check", "application/json", requestBody)
	if err != nil {
		log.Error("cannot send send POST request to imageOps Check, with filename: " + saveFileName)
		c.writeErrorResponse("cannot send request to imagesOps",util.StatusNotFound)
		return CheckResponse{}, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	var checkResponse CheckResponse
	err = json.Unmarshal(body, &checkResponse)
	if err != nil {
		log.Error(util.FailedToUnmarshal)
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
		return CheckResponse{}, err
	}
	return checkResponse, nil
}
