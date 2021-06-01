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
	"errors"
	"fileSystem/models"
	"fileSystem/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

type UploadController struct {
	BaseController
}

func createDirectory(dir string) error { //make dir if path doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.New("failed to create directory")
		}
	}
	return nil
}

func createImageID() string {
	uuId := uuid.NewV4()
	return strings.Replace(uuId.String(), "-", "", -1)
}

func (c *UploadController) insertOrUpdateFileRecord(imageId, fileName, userId, saveFileName, storageMedium string) error {

	fileRecord := &models.ImageDB{
		ImageId: imageId,
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
func (c *UploadController) getStorageMedium(priority string) string {
	switch {
	case priority == "A":
		return "huaweiCloud"

	case priority == "B":
		return "Azure"

	default:
		defaultPath := util.LocalStoragePath
		return defaultPath
	}
}

// @Title Post
// @Description upload file
// @Param   priority     string  true   "priority "
// @Param   saveFilename 	string  	true   "file"   eg.9c73996089944709bad8efa7f532aebe1.zip
func (c *UploadController) saveByPriority(priority string, saveFilename string) error{
	switch {
	case priority == "A":
		return errors.New("sorry, this storage medium is not supported right now")

	default:
		defaultPath := util.LocalStoragePath

		err := createDirectory(defaultPath)
		if err != nil {
			log.Error("failed to create file path" + defaultPath)
			return err
		}

		err = c.SaveToFile(util.FormFile, defaultPath+saveFilename)
		if err != nil {
			c.writeErrorResponse("fail to upload package", util.StatusInternalServerError)
			return err
		} else {
			log.Info("save file to " + defaultPath)
			return nil
		}
	}
}

// @Title Get
// @Description test connection is ok or not
// @Success 200 ok
// @Failure 400 bad request
// @router /imagemanagement/v1/upload [get]
func (c *UploadController) Get() {
	log.Info("Upload get request received.")
	c.Ctx.WriteString("Upload get request received.")
}

// @Title Post
// @Description upload file
// @Param   usrId       form-data 	string	true   "usrId"
// @Param   priority    form-data   string  true   "priority "
// @Param   file        form-data 	file	true   "file"
// @Success 200 ok
// @Failure 400 bad request
// @router /imagemanagement/v1/upload [post]
func (c *UploadController) Post() {
	log.Info("Upload post request received.")

	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}

	c.displayReceivedMsg(clientIp)

	file, head, err := c.GetFile("file")
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, "Upload package file error")
		return
	}

	err = util.ValidateFileExtensionZip(head.Filename)
	if err != nil || len(head.Filename) > util.MaxFileNameSize {
		c.HandleLoggingForError(clientIp, util.BadRequest,
			"File shouldn't contains any extension or filename is larger than max size")
		return
	}

	err = util.ValidateFileSize(head.Size, util.MaxAppPackageFile)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, "File size is larger than max size")
		return
	}
	defer file.Close()

	filename := head.Filename //original name for file

	userId := c.GetString(util.UserId)
	priority := c.GetString(util.Priority)

	//create imageId, fileName, uploadTime, userId
	imageId := createImageID()

	//get a storage medium to let fe know
	storageMedium := c.getStorageMedium(priority)

	saveFileName := imageId + filename     //9c73996089944709bad8efa7f532aebe+1.zip

	err = c.saveByPriority(priority, saveFileName)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to upload package")
		return
	}

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
