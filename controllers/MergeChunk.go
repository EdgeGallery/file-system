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
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//ChunkUploadController  Define the controller to control upload
type MergeChunkController struct {
	BaseController
}

func (c *MergeChunkController) insertOrUpdateFileRecord(imageId, fileName, userId, saveFileName, storageMedium, requestIdCheck string) error {
	fileRecord := &models.ImageDB{
		ImageId:        imageId,
		FileName:       fileName,
		UserId:         userId,
		SaveFileName:   saveFileName,
		StorageMedium:  storageMedium,
		RequestIdCheck: requestIdCheck,
	}
	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")
	if err != nil && err.Error() != util.LastInsertIdNotSupported {
		log.Error(util.FailToRecordToDB)
		return err
	}
	log.Info(util.FileRecord, fileRecord)
	return nil
}

func (c *MergeChunkController) insertOrUpdateCheckRecord(imageId, fileName, userId, storageMedium, saveFileName string, slimStatus int, checkStatusResponse CheckStatusResponse) error {
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

//add more storage logic here
func (c *MergeChunkController) getStorageMedium(priority string) string {
	switch {
	case priority == "A":
		return "huaweiCloud"
	case priority == "B":
		return "Azure"
	default:
		defaultPath := util.LocalStoragePath // "/usr/app/vmImage/"
		return defaultPath
	}
}

// @Title Get
// @Description test connection is ok or not
// @Success 200 ok
// @Failure 400 bad request
// @router "/image-management/v1/images/merge [get]
func (c *MergeChunkController) Get() {
	log.Info("Merge get request received.")
	c.Ctx.WriteString("Merge get request received.")
}

// @Title Post
// @Description merge chunk file
// @Param   identifier  form-data 	string	true   "identifier"
// @Param   filename    form-data 	string	true   "filename"
// @Param   userId      form-data 	string 	true   "userId"
// @Param   priority    form-data   string  true   "priority"
// @Success 200 ok
// @Failure 400 bad request
// @router "/image-management/v1/images/merge [post]
func (c *MergeChunkController) Post() {
	log.Info("Upload post request received.")
	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}
	c.displayReceivedMsg(clientIp)

	userId := c.GetString(util.UserId)
	identifier := c.GetString(util.Identifier)
	filename := c.GetString(util.FileName)

	err = util.ValidateFileExtension(filename)
	if err != nil || len(filename) > util.MaxFileNameSize {
		c.HandleLoggingForError(clientIp, util.BadRequest,
			"File should only be image file or filename is larger than max size")
		return
	}

	priority := c.GetString(util.Priority)
	//create imageId, fileName, uploadTime, userId
	imageId := CreateImageID()
	//get a storage medium to let fe know
	storageMedium := c.getStorageMedium(priority)
	saveFilePath := storageMedium + imageId + filename //   app/vmImages/identifier/xx.zip
	file, err := os.OpenFile(saveFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to find the previous file path")
		return
	}
	files, err := ioutil.ReadDir(storageMedium + identifier + "/")
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to find the file path")
		return
	}
	totalChunksNum := len(files) //total number of chunks
	for i := 1; i <= totalChunksNum; i++ {
		tmpFilePath := storageMedium + identifier + "/" + strconv.Itoa(i) + ".part"
		f, err := os.OpenFile(tmpFilePath, os.O_RDONLY, os.ModePerm)
		if err != nil {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to open the part file in path")
			return
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to read the part file in path")
			return
		}
		file.Write(b)
		f.Close()
		err = os.Remove(tmpFilePath)
		if err != nil {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to delete the part file in path")
			return
		}
	}
	file.Close()

	saveFileName := imageId + filename
	if filepath.Ext(filename) == ".zip" {
		filenameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
		decompressFilePath := saveFilePath
		arr, err := DeCompress(decompressFilePath, storageMedium+filenameWithoutExt)
		if err != nil {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailedToDecompress)
			return
		}
		originalName := subString(arr[0], strings.LastIndex(arr[0], "/")+1, len(arr[0]))
		saveFileName = imageId + originalName
		srcFileName := arr[0]
		dstFileName := storageMedium + saveFileName
		_, err = CopyFile(srcFileName, dstFileName)
		if err != nil {
			log.Error("when decompress, failed to copy file")
			return
		}
		err = os.RemoveAll(storageMedium + filenameWithoutExt + "/")
		if err != nil {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to delete tmp file package in vm")
			return
		}
		err = os.Remove(decompressFilePath)
		if err != nil {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to delete tmp zip file in vm")
			return
		}
		filename = originalName
	}

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
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "cannot send request to imagesOps")
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	var checkResponse CheckResponse
	err = json.Unmarshal(body, &checkResponse)
	if err != nil {
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
	}
	status := checkResponse.Status
	msg := checkResponse.Msg
	requestIdCheck := checkResponse.RequestId

	err = c.insertOrUpdateFileRecord(imageId, filename, userId, saveFileName, storageMedium, requestIdCheck)
	if err != nil {
		log.Error(util.FailedToInsertDataToDB)
		return
	}
	//delete the emp file path
	err = os.RemoveAll(storageMedium + identifier + "/")
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to delete part file in vm")
		return
	}

	slimStatus := "0" //默认未瘦身
	uploadResp, err := json.Marshal(map[string]interface{}{
		"imageId":       imageId,
		"fileName":      filename,
		"uploadTime":    time.Now().Format("2006-01-02 15:04:05"),
		"userId":        userId,
		"storageMedium": storageMedium,
		"slimStatus":    slimStatus,
		"checkStatus":   status,
		"msg":           msg,
	})
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return upload details")
		return
	}
	_, _ = c.Ctx.ResponseWriter.Write(uploadResp)

	go c.helper(requestIdCheck, clientIp, client, imageId, filename, userId, storageMedium, saveFileName)

}

func (c *MergeChunkController) helper(requestIdCheck string, clientIp string, client *http.Client, imageId string, filename string, userId string, storageMedium string, saveFileName string) {
	//此时瘦身结束，查看Check Response详情
	isCheckFinished := false
	checkTimes := 60
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
			time.Sleep(time.Duration(30) * time.Second)
			continue
		} else {      //check completed
			isCheckFinished = true
			err = c.insertOrUpdateCheckRecord(imageId, filename, userId, storageMedium, saveFileName, 0, checkStatusResponse)
			if err != nil {
				log.Error(util.FailedToInsertDataToDB)
				c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailToInsertRequestCheck)
				return
			}
		}
	}
}
