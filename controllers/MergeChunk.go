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
	"strconv"
	"time"
)

//ChunkUploadController  Define the controller to control upload
type MergeChunkController struct {
	BaseController
}

func (c *MergeChunkController) insertOrUpdateFileRecord(imageId, fileName, userId, saveFileName, storageMedium string) error {

	fileRecord := &models.ImageDB{
		ImageId:       imageId,
		FileName:      fileName,
		UserId:        userId,
		SaveFileName:  saveFileName,
		StorageMedium: storageMedium,
	}

	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")

	if err != nil && err.Error() != "LastInsertId is not supported by this driver" {
		log.Error("Failed to save file record to database.")
		return err
	}

	log.Info("Add file record: %+v", fileRecord)
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
// @router "/image-management/v1/images [get]
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
// @router "/image-management/v1/images [post]
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
	filename := c.GetString(util.FileName) //xxxxxxx.zip
	priority := c.GetString(util.Priority)

	//create imageId, fileName, uploadTime, userId
	imageId := createImageID()

	//get a storage medium to let fe know
	storageMedium := c.getStorageMedium(priority)

	saveFilePath := storageMedium + imageId + filename //   app/vmImages/identifier/xx.zip

	file, err := os.OpenFile(saveFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to find the previous file path")
		return
	}

	//files, err := filepath.Glob()
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
	}

	saveFileName := imageId + filename

	err = c.insertOrUpdateFileRecord(imageId, filename, userId, saveFileName, storageMedium)
	if err != nil {
		log.Error("fail to insert imageID, filename, userID to database")
		return
	}

	uploadResp, err := json.Marshal(map[string]string{
		"imageId":       imageId,
		"fileName":      filename,
		"uploadTime":    time.Now().Format("2006-01-02 15:04:05"),
		"userId":        userId,
		"storageMedium": storageMedium,
	})
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return upload details")
		return
	}
	_, _ = c.Ctx.ResponseWriter.Write(uploadResp)

}
