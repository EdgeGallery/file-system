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
	"encoding/json"
	"fileSystem/models"
	"fileSystem/util"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
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

//add more storage logic here
func (c *MergeChunkController) GetStorageMedium(priority string) string {
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
	log.Info("Merge post request received.")
	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}
	c.displayReceivedMsg(clientIp)

	//TODO: 校验userId、priority 加一个校验
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
	storageMedium := c.GetStorageMedium(priority)
	saveFilePath := storageMedium + imageId + filename //   app/vmImages/identifier/xx.zip
	file, err := os.OpenFile(saveFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to find the previous file path")
		return
	}
	log.Info("open file from "+ saveFilePath)
	files, err := ioutil.ReadDir(storageMedium + identifier + "/")
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to find the file path")
		return
	}
	totalChunksNum := len(files) //total number of chunks
	log.Info("The total file chunk number is "+ strconv.Itoa(totalChunksNum))
	for i := 1; i <= totalChunksNum; i++ {
		log.Info("loading " + strconv.Itoa(i)+"th file chunk")
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
		log.Info( strconv.Itoa(i)+"th file chunk merge success")
	}
	file.Close()

	log.Info("Chunk files merge finished.")
	saveFileName := imageId + filename
	if filepath.Ext(filename) == ".zip" {
		log.Info("begin to compress the merged file to zip file")
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
			log.Error("when compressing, failed to copy file")
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
		log.Info("compress the merged file to zip file success")
	}

	log.Info("begin to request to imageOps check with POST")
	checkResponse, err := c.PostToCheck(saveFileName)
	if err != nil {
		log.Error("cannot send send POST request to imageOps Check, with filename: " + saveFileName)
		c.writeErrorResponse("cannot send request to imagesOps", util.StatusNotFound)
		return
	}
	status := checkResponse.Status
	msg := checkResponse.Msg
	requestIdCheck := checkResponse.RequestId
	log.Info("get Check requestId from imageOps with "+ requestIdCheck)

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
	log.Info("delete temporary file path from: "+ storageMedium + identifier + "/")

	uploadResp, err := json.Marshal(map[string]interface{}{
		"imageId":       imageId,
		"fileName":      filename,
		"uploadTime":    time.Now().Format("2006-01-02 15:04:05"),
		"userId":        userId,
		"storageMedium": storageMedium,
		"slimStatus":    util.UnSlimmed,
		"checkStatus":   status,
		"msg":           msg,
	})
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return upload details")
		return
	}
	_, _ = c.Ctx.ResponseWriter.Write(uploadResp)

	log.Info("begin to request to imageOps check with GET")
	time.Sleep(time.Duration(5) * time.Second)
	go c.CronGetCheck(requestIdCheck, imageId,  filename, userId, storageMedium, saveFileName)
}


