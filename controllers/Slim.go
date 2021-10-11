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
	ImageEndOffset string `json:"image-end-offset"`
	CheckErrors    string `json:"check-errors"`
	Format         string `json:"format"`
	Filename       string `json:"filename"`
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
	Rate   float64 `json:"rate"`
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

func (c *SlimController) insertOrUpdatePostRecord(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int, requestId string) error {
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
	if err != nil && err.Error() != "LastInsertId is not supported by this driver" {
		log.Error("Failed to save file record to database.")
		return err
	}
	log.Info("Add file record: %+v", fileRecord)
	return nil
}

func (c *SlimController) insertOrUpdateCheckPostRecord(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int, requestId string) error {
	fileRecord := &models.ImageDB{
		ImageId:        imageId,
		FileName:       fileName,
		UserId:         userId,
		StorageMedium:  storageMedium,
		SaveFileName:   saveFileName,
		SlimStatus:     slimStatus,
		RequestIdCheck: requestId,
	}
	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")
	if err != nil && err.Error() != "LastInsertId is not supported by this driver" {
		log.Error("Failed to save file record to database.")
		return err
	}
	log.Info("Add file record: %+v", fileRecord)
	return nil
}

func (c *SlimController) insertOrUpdateCheckRecord(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int, checkStatusResponse CheckStatusResponse) error {
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
	if err != nil && err.Error() != "LastInsertId is not supported by this driver" {
		log.Error("Failed to save file record to database.")
		return err
	}
	log.Info("Add file record: %+v", fileRecord)
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
	if imageFileDb.SlimStatus == 1 { //此时镜像正在瘦身 [0,1,2,3]  未瘦身/瘦身中/成功/失败
		log.Info("the image file is being slimmed. No need to slim again.")
		c.Ctx.WriteString("the image file is being slimmed. No need to slim again.")
		return
	}
	if imageFileDb.SlimStatus == 2 { //此时镜像已经瘦身
		log.Info("the image file has already been slimmed. No need to slim again. Pls request to check directly")
		c.Ctx.WriteString("the image file has already been slimmed. No need to slim again.Pls request to check directly")
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
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	var compressRes CompressResult
	err = json.Unmarshal(body, &compressRes)
	if err != nil {
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
	}

	requestIdCompress := compressRes.RequestId
	responseStatus := compressRes.Status //0:compress in progress  1: compress failed
	if responseStatus == 0 {
		c.Ctx.WriteString("compress in progress")
		err = c.insertOrUpdatePostRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, 1, requestIdCompress) //[0,1,2,3]  未瘦身/瘦身中/成功/失败
		if err != nil {
			log.Error("fail to insert imageId,slimStatus,requestId to database")
			return
		}
	} else if responseStatus == 1 {
		err = c.insertOrUpdatePostRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, 3, requestIdCompress) //[0,1,2,3]  未瘦身/瘦身中/成功/失败
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "imageOps compress failed")
		return
	}
	go func() {
		//此时正在瘦身
		var requestIdCheck string
		isCompressFinished := false
		checkTimes := 60
		for !isCompressFinished && checkTimes > 0 {
			checkTimes--
			response, err := client.Get("http://localhost:5000/api/v1/vmimage/compress/" + requestIdCompress)
			if err != nil {
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to request vmimage compress check")
				return
			}
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			var compressStatusResponse CompressStatusResponse
			err = json.Unmarshal(body, &compressStatusResponse)
			if err != nil {
				log.Error("fail to request http://localhost:5000/api/v1/vmimage/compress/" + requestIdCompress)
				c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
				return
			}
			if compressStatusResponse.Status == 1 { //compress in progress
				time.Sleep(time.Duration(30) * time.Second)
				continue
			} else if compressStatusResponse.Status == 0 { //compress finished
				//瘦身成功后，及时更新数据库
				//err = c.insertOrUpdatePostRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, 1, requestIdCompress) // slimStatus == 2 瘦身成功
				isCompressFinished = true
				saveFileName := imageFileDb.SaveFileName
				var formConfigMap map[string]string
				formConfigMap = make(map[string]string)
				formConfigMap["inputImageName"] = "compressed" + saveFileName
				requestJson, _ := json.Marshal(formConfigMap)
				requestBody := bytes.NewReader(requestJson)
				response, err := client.Post("http://localhost:5000/api/v1/vmimage/check", "application/json", requestBody)
				if err != nil {
					c.HandleLoggingForError(clientIp, util.StatusNotFound, "Slim POST cannot send request to imagesOps")
					return
				}
				defer response.Body.Close()
				body, err := ioutil.ReadAll(response.Body)
				var checkResponse CheckResponse
				err = json.Unmarshal(body, &checkResponse)
				if err != nil {
					c.writeErrorResponse("Slim POST to image check failed to unmarshal request", util.BadRequest)
				}
				requestIdCheck = checkResponse.RequestId
				if len(requestIdCheck) == 0 {
					c.Ctx.WriteString("check requestId is empty, check if imageOps is ok")
					return
				}
				err = c.insertOrUpdateCheckPostRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, 1, requestIdCheck) // slimStatus == 1 瘦身中
				break
			} else if compressStatusResponse.Status == 2 { //compress failed
				isCompressFinished = true
				err = c.insertOrUpdatePostRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, 3, requestIdCompress) //[0,1,2,3]  未瘦身/瘦身中/成功/失败
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "imageOps compress failed")
				return
			}
		}

		//此时瘦身结束，查看Check Response详情
		isCheckFinished := false
		checkTimes = 180  //30 MinS
		for !isCheckFinished && checkTimes > 0 {
			checkTimes--
			if len(requestIdCheck) == 0 {
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "after POST check to imageOps, check requestId is till empty")
				return
			}
			responseCheck, err := client.Get("http://localhost:5000/api/v1/vmimage/check/" + requestIdCheck)
			if err != nil {
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to request imageOps check")
				return
			}
			defer responseCheck.Body.Close()
			bodyCheck, err := ioutil.ReadAll(responseCheck.Body)
			var checkStatusResponse CheckStatusResponse
			err = json.Unmarshal(bodyCheck, &checkStatusResponse)
			if err != nil {
				c.writeErrorResponse("Slim GET to image check failed to unmarshal request", util.BadRequest)
				return
			}
			if checkStatusResponse.Status == 4 { // check in progress
				time.Sleep(time.Duration(10) * time.Second)
				continue
			} else if checkStatusResponse.Status == 0 { //check completed
				isCheckFinished = true
				err = c.insertOrUpdateCheckRecord(imageId, imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName,2, checkStatusResponse)
				if err != nil {
					log.Error("fail to insert imageID, filename, userID to database")
					c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to insert request imageOps check to db")
					return
				}
			} else {
				isCheckFinished = true
				err = c.insertOrUpdateCheckRecord(imageId,imageFileDb.FileName, imageFileDb.UserId, imageFileDb.StorageMedium, imageFileDb.SaveFileName, 3, checkStatusResponse)
				if err != nil {
					log.Error("fail to insert imageID, filename, userID to database")
					c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to insert request imageOps check to db")
					return
				}
			}
		}
	}()
}