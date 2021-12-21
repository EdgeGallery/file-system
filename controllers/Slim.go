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

package controllers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fileSystem/models"
	"fileSystem/util"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type SlimController struct {
	BaseController
}

// /vmimage/check post response
// /vmimage/compress post response
type CompressResult struct {
	Status    int    `json:"status"`
	Msg       string `json:"msg"`
	RequestId string `json:"requestId"`
}

type ImageInfo struct {
	ImageEndOffset string  `json:"image-end-offset"`
	CheckErrors    string  `json:"check-errors"`
	Format         string  `json:"format"`
	Filename       string  `json:"filename"`
	VirtualSize    float32 `json:"virtual_size"`
	DiskSize       string  `json:"disk_size"`
}

type CheckInfo struct {
	Checksum         string    `json:"checksum"`
	CheckResult      int       `json:"checkResult"`
	ImageInformation ImageInfo `json:"imageInfo"`
}

// /vmimage/check/requestId get response
type CheckStatusResponse struct {
	Status           int       `json:"status"`
	Msg              string    `json:"msg"`
	CheckInformation CheckInfo `json:"checkInfo"`
}

// /vmimage/compress/requestId get response
type CompressStatusResponse struct {
	Status int     `json:"status"`
	Msg    string  `json:"msg"`
	Rate   float32 `json:"rate"`
}

// @Title PathCheck
// @Description check file in path is existed or not
// @Param   Source Zip File Path    string
func (c *SlimController) PathCheck(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func (c *SlimController) insertOrUpdateCompressPostRecord(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int, requestId string) error {
	fileRecord := &models.ImageDB{
		FileName:          fileName,
		UserId:            userId,
		StorageMedium:     storageMedium,
		SaveFileName:      saveFileName,
		ImageId:           imageId,
		SlimStatus:        slimStatus,
		RequestIdCompress: requestId,
	}
	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")
	if err != nil && err.Error() != util.LastInsertIdNotSupported {
		log.Error(util.FailToRecordToDB)
		return err
	}
	log.Info(util.FileRecord, fileRecord)
	return nil
}

func (c *SlimController) insertOrUpdateCheckPostRecordAfterCompress(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int, requestId string, compressStatusResponse CompressStatusResponse) error {
	fileRecord := &models.ImageDB{
		ImageId:        imageId,
		FileName:       fileName,
		UserId:         userId,
		StorageMedium:  storageMedium,
		SaveFileName:   saveFileName,
		SlimStatus:     slimStatus,
		RequestIdCheck: requestId,
		//Compress进度参数
		CompressStatus: compressStatusResponse.Status,
		CompressMsg:    compressStatusResponse.Msg,
		CompressRate:   compressStatusResponse.Rate,
	}
	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")
	if err != nil && err.Error() != util.LastInsertIdNotSupported {
		log.Error(util.FailToRecordToDB)
		return err
	}
	log.Info(util.FileRecord, fileRecord)
	return nil
}

func (c *SlimController) insertOrUpdateCommonRecord(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int) error {
	fileRecord := &models.ImageDB{
		ImageId:       imageId,
		FileName:      fileName,
		UserId:        userId,
		StorageMedium: storageMedium,
		SaveFileName:  saveFileName,
		SlimStatus:    slimStatus,
	}
	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")
	if err != nil && err.Error() != util.LastInsertIdNotSupported {
		log.Error(util.FailToRecordToDB)
		return err
	}
	log.Info(util.FileRecord, fileRecord)
	return nil
}

func (c *SlimController) insertOrUpdateCompressGetRecord(imageId, fileName, userId, storageMedium, saveFileName string, requestId string, slimStatus int, compressStatusResponse CompressStatusResponse) error {
	log.Info("begin to insert compress rate, status, msg to database.")
	fileRecord := &models.ImageDB{
		ImageId:           imageId,
		FileName:          fileName,
		UserId:            userId,
		StorageMedium:     storageMedium,
		SaveFileName:      saveFileName,
		RequestIdCompress: requestId,
		SlimStatus:        slimStatus,
		//Compress进度参数
		CompressStatus: compressStatusResponse.Status,
		CompressMsg:    compressStatusResponse.Msg,
		CompressRate:   compressStatusResponse.Rate,
	}
	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")
	if err != nil && err.Error() != util.LastInsertIdNotSupported {
		log.Error(util.FailToRecordToDB)
		return err
	}
	log.Info(util.FileRecord, fileRecord)
	return nil
}

func (c *SlimController) insertOrUpdateCheckRecordAfterCompress(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int, checkStatusResponse CheckStatusResponse, compressStatusResponse CompressStatusResponse) error {
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
		VirtualSize:    checkStatusResponse.CheckInformation.ImageInformation.VirtualSize,
		DiskSize:       checkStatusResponse.CheckInformation.ImageInformation.DiskSize,
		//Compress进度参数
		CompressStatus: compressStatusResponse.Status,
		CompressMsg:    compressStatusResponse.Msg,
		CompressRate:   compressStatusResponse.Rate,
	}
	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")
	if err != nil && err.Error() != util.LastInsertIdNotSupported {
		log.Error(util.FailToRecordToDB)
		return err
	}
	log.Info(util.FileRecord, fileRecord)
	return nil
}

// @Title Post
// @Description perform image slim operation
// @Param	imageId 	string
// @Success 200 ok
// @Failure 400 bad request
// @router "/image-management/v1/images/:imageId/action/slim [post]
func (c *SlimController) Post() {
	log.Info("image slim Post request received.")
	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}
	c.displayReceivedMsg(clientIp)
	var imageFileDb models.ImageDB
	imageId := c.Ctx.Input.Param(":imageId")
	_, err = c.Db.QueryTable("image_d_b", &imageFileDb, "image_id__exact", imageId)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "fail to query database")
		return
	}
	if imageFileDb.SlimStatus == util.Slimming { //此时镜像正在瘦身 [0,1,2,3]  未瘦身/瘦身中/成功/失败
		log.Info(util.ImageSlimming)
		c.Ctx.WriteString(util.ImageSlimming)
		return
	}

	if imageFileDb.SlimStatus == util.SlimmedSuccess { //此时镜像已经瘦身
		log.Info(util.ImageSlimmed)
		c.Ctx.WriteString(util.ImageSlimmed)
		return
	}

	if imageFileDb.CheckStatus == util.CheckUnsupportedType { //镜像格式不支持瘦身
		log.Info(util.TypeNotSupport)
		err := c.insertOrUpdateCommonRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.SlimFailed) //[0,1,2,3]  未瘦身/瘦身中/成功/失败
		if err != nil {
			log.Error(util.FailedToInsertDataToDB)
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailToInsertRequestCheck)
		}
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.TypeNotSupport)
		return
	}

	filePath := imageFileDb.StorageMedium
	saveFilename := imageFileDb.SaveFileName
	if !c.PathCheck(filePath) {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "file path doesn't exist")
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	compressRes, done := c.postToCompress(saveFilename, client, clientIp)
	if done {
		return
	}
	requestIdCompress := compressRes.RequestId
	responseStatus := compressRes.Status //0:compress in progress  1: compress failed
	if responseStatus == util.PostCompressInProgress {
		c.Ctx.WriteString("compress in progress")
		err = c.insertOrUpdateCompressPostRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.Slimming, requestIdCompress) //[0,1,2,3]  未瘦身/瘦身中/成功/失败
		if err != nil {
			log.Error(util.FailedToInsertDataToDB)
			return
		}
	} else if responseStatus == util.PostCompressFailed {
		c.Ctx.WriteString("compress failed")
		err = c.insertOrUpdateCompressPostRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.SlimFailed, requestIdCompress) //[0,1,2,3]  未瘦身/瘦身中/成功/失败
		if err != nil {
			log.Error(util.FailedToInsertDataToDB)
			return
		}
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "Before asyCall imageOps, imageOps compress failed")
		return
	} else { // 逃生通道
		errMsg := "Compress response error, return code has not defined:" + strconv.Itoa(responseStatus)
		c.Ctx.WriteString(errMsg)
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, errMsg)
		return
	}
	time.Sleep(time.Duration(5) * time.Second)
	//异步调用
	go c.asyCallImageOps(client, requestIdCompress, clientIp, imageFileDb, imageId)

}

func (c *SlimController) asyCallImageOps(client *http.Client, requestIdCompress string, clientIp string, imageFileDb models.ImageDB, imageId string) {

	//此时正在瘦身
	var requestIdCheck string
	var compressInfo CompressStatusResponse

	isCompressFinished := false
	checkTimes := 60
	for !isCompressFinished && checkTimes > 0 {
		checkTimes--
		compressStatusResponse, isFailed := c.getToCompress(client, requestIdCompress, clientIp)
		compressInfo = compressStatusResponse
		if isFailed == true {
			return
		}
		if compressStatusResponse.Status == util.CompressInProgress { //compress in progress

			log.Info("imageOps is compressing, the imageId is " + imageId)
			err := c.insertOrUpdateCompressGetRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, requestIdCompress, util.Slimming, compressStatusResponse)
			if err != nil {
				log.Error("When imageOps compressing, Compress GET record cannot be inserted to database")
				c.writeErrorResponse(util.FailToRecordToDB, util.BadRequest)
				return
			}
			time.Sleep(time.Duration(30) * time.Second)
			continue
		} else if compressStatusResponse.Status == util.CompressCompleted { //compress completed

			log.Info("imageOps compress finished, the imageId is " + imageId)
			err := c.insertOrUpdateCompressGetRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, requestIdCompress, util.Slimming, compressStatusResponse)
			if err != nil {
				log.Error("When imageOps compress finished, Compress GET record cannot be inserted to database")
				c.writeErrorResponse(util.FailToRecordToDB, util.BadRequest)
				return
			}

			isCompressFinished = true
			// begin to check slimmed image
			checkResponse, isFailed := c.postToCheck(client, imageFileDb, clientIp)
			if isFailed == true {
				return
			}
			requestIdCheck = checkResponse.RequestId
			if len(requestIdCheck) == 0 {
				log.Error("check requestId is empty, check if imageOps is ok")
				c.Ctx.WriteString("check requestId is empty, check if imageOps is ok")
				return
			}
			err = c.insertOrUpdateCheckPostRecordAfterCompress(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.Slimming, requestIdCheck, compressStatusResponse) // slimStatus == 1 瘦身中
			if err != nil {
				c.writeErrorResponse(util.FailToRecordToDB, util.BadRequest)
				return
			}
			break
		} else if compressStatusResponse.Status == util.CompressFailed { //compress failed

			isCompressFinished = true
			log.Error("imageOps compress failed, the imageId is " + imageId)

			err := c.insertOrUpdateCheckPostRecordAfterCompress(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.SlimFailed, requestIdCheck, compressStatusResponse)
			if err != nil {
				log.Error("When imageOps compress failed, Compress GET record cannot be inserted to database")
				c.writeErrorResponse(util.FailToRecordToDB, util.BadRequest)
				return
			}

			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "After compress, imageOps compress failed")
			return
		} else if compressStatusResponse.Status == util.CompressNoEnoughSpace { // compress exit, since no space left

			isCompressFinished = true
			log.Error(util.SlimExitNoSpace)

			err := c.insertOrUpdateCheckPostRecordAfterCompress(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.SlimFailed, requestIdCheck, compressStatusResponse)
			if err != nil {
				log.Error("When imageOps  compress exit since no space left, Compress GET record cannot be inserted to database")
				c.writeErrorResponse(util.FailToRecordToDB, util.BadRequest)
				return
			}

			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.SlimExitNoSpace)
			return
		} else if compressStatusResponse.Status == util.CompressTimeOut { // compress exit, since time out

			log.Error(util.CompressTimeOutMsg)
			isCompressFinished = true
			err := c.insertOrUpdateCheckPostRecordAfterCompress(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.SlimFailed, requestIdCheck, compressStatusResponse)
			if err != nil {
				log.Error("When imageOps  compress exit since time out, Compress GET record cannot be inserted to database")
				c.writeErrorResponse(util.FailToRecordToDB, util.BadRequest)
				return
			}
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.CompressTimeOutMsg)
			return
		} else { // 增加逃生通道
			log.Error(util.UnknownCompressStatus + ":" + strconv.Itoa(compressStatusResponse.Status))
			isCompressFinished = true
			err := c.insertOrUpdateCheckPostRecordAfterCompress(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.SlimFailed, requestIdCheck, compressStatusResponse)
			if err != nil {
				log.Error(util.FailedToInsertDataToDB)
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailToInsertRequestCheck)
			}
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.UnknownCompressStatus)
			return
		}
	}

	time.Sleep(time.Duration(5) * time.Second)
	//此时瘦身结束，查看Check Response详情
	isCheckFinished := false
	checkTimes = 180 //30 MinS

	for !isCheckFinished && checkTimes > 0 {

		checkTimes--
		if len(requestIdCheck) == 0 {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "after POST check to imageOps, check requestId is still empty")
			return
		}

		checkStatusResponse, undone := c.getToCheck(client, requestIdCheck, clientIp)
		if undone {
			return
		}

		if checkStatusResponse.Status == util.CheckInProgress { // check in progress
			time.Sleep(time.Duration(10) * time.Second)
			err := c.insertOrUpdateCheckRecordAfterCompress(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.Slimming, checkStatusResponse, compressInfo)
			if err != nil {
				log.Error(util.FailedToInsertDataToDB)
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailToInsertRequestCheck)
				return
			}
			continue
		} else if checkStatusResponse.Status == util.CheckCompleted { //check completed
			isCheckFinished = true

			err := c.insertOrUpdateCheckRecordAfterCompress(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.SlimmedSuccess, checkStatusResponse, compressInfo)
			if err != nil {
				log.Error(util.FailedToInsertDataToDB)
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailToInsertRequestCheck)
				return
			}
		} else {
			isCheckFinished = true
			err := c.insertOrUpdateCheckRecordAfterCompress(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, util.SlimFailed, checkStatusResponse, compressInfo)
			if err != nil {
				log.Error(util.FailedToInsertDataToDB)
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailToInsertRequestCheck)
				return
			}
		}
	}
}

func (c *SlimController) getToCheck(client *http.Client, requestIdCheck string, clientIp string) (CheckStatusResponse, bool) {
	responseCheck, err := client.Get("http://localhost:5000/api/v1/vmimage/check/" + requestIdCheck)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to request imageOps check")
		return CheckStatusResponse{}, true
	}
	defer responseCheck.Body.Close()
	bodyCheck, err := ioutil.ReadAll(responseCheck.Body)
	var checkStatusResponse CheckStatusResponse
	err = json.Unmarshal(bodyCheck, &checkStatusResponse)
	if err != nil {
		c.writeErrorResponse("Slim GET to image check failed to unmarshal request", util.BadRequest)
		return CheckStatusResponse{}, true
	}
	return checkStatusResponse, false
}

func (c *SlimController) postToCheck(client *http.Client, imageFileDb models.ImageDB, clientIp string) (CheckResponse, bool) {
	saveFileName := imageFileDb.SaveFileName
	var formConfigMap map[string]string
	formConfigMap = make(map[string]string)
	formConfigMap["inputImageName"] = "compressed" + saveFileName
	requestJson, _ := json.Marshal(formConfigMap)
	requestBody := bytes.NewReader(requestJson)
	response, err := client.Post("http://localhost:5000/api/v1/vmimage/check", "application/json", requestBody)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "Slim POST cannot send request to imagesOps")
		return CheckResponse{}, true
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	var checkResponse CheckResponse
	err = json.Unmarshal(body, &checkResponse)
	if err != nil {
		c.writeErrorResponse("Slim POST to image check failed to unmarshal request", util.BadRequest)
	}
	return checkResponse, false
}

func (c *SlimController) getToCompress(client *http.Client, requestIdCompress string, clientIp string) (CompressStatusResponse, bool) {
	response, err := client.Get("http://localhost:5000/api/v1/vmimage/compress/" + requestIdCompress)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to request vmimage compress check")
		return CompressStatusResponse{}, true
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to read response body")
		return CompressStatusResponse{}, true
	}
	var compressStatusResponse CompressStatusResponse
	err = json.Unmarshal(body, &compressStatusResponse)
	if err != nil {
		log.Error("fail to request http://localhost:5000/api/v1/vmimage/compress/" + requestIdCompress)
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
		return CompressStatusResponse{}, true
	}
	return compressStatusResponse, false
}

func (c *SlimController) postToCompress(saveFilename string, client *http.Client, clientIp string) (CompressResult, bool) {
	var formConfigMap map[string]string
	formConfigMap = make(map[string]string)
	formConfigMap["inputImageName"] = saveFilename
	formConfigMap["outputImageName"] = "compressed" + saveFilename

	requestJson, _ := json.Marshal(formConfigMap)
	requestBody := bytes.NewReader(requestJson)

	//imageops/api/v1/vmimage/compress
	response, err := client.Post("http://localhost:5000/api/v1/vmimage/compress", "application/json", requestBody)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "cannot send request to imagesOps")
		return CompressResult{}, true
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	var compressRes CompressResult
	err = json.Unmarshal(body, &compressRes)
	if err != nil {
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
		return CompressResult{}, true
	}
	return compressRes, false
}
