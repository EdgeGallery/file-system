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

/*var (
	PackageFolderPath string //写到enev
	//PackageFolderPath = "static/"
)*/

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

func (this *UploadController) insertOrUpdateFileRecord(imageId, fileName, userId, saveFileName, storageMedium, url string) error {

	fileRecord := &models.ImageDB{
		ImageId:       imageId,
		FileName:      fileName,
		UserId:        userId,
		SaveFileName:  saveFileName,
		StorageMedium: storageMedium,
		Url:           url}

	err := this.Db.InsertOrUpdateData(fileRecord, "image_id")

	if err != nil && err.Error() != "LastInsertId is not supported by this driver" {
		log.Error("Failed to save file record to database.")
		return err
	}

	log.Info("Add file record: %+v", fileRecord)
	return nil
}

//add more storage logic here
func (this *UploadController) getStorageMedium(priority string) string {
	switch {
	case priority == "1":
		return "huaweiCloud"

	case priority == "2":
		return "AWS"

	default:
		defaultPath := "/usr/vmImage/"
		return defaultPath
	}
}

// @Title Get
// @Description test connection is ok or not
// @Success 200 ok
// @Failure 400 bad request
// @router /imagemanagement/v1/upload [get]
func (this *UploadController) Get() {
	log.Info("Upload get request received.")
	this.Ctx.WriteString("Upload get request received.")
}

// @Title Post
// @Description upload file
// @Param   usrId       form-data 	string	true   "usrId"
// @Param   priority    form-data   string  true   "priority "
// @Param   file        form-data 	file	true   "file"
// @Success 200 ok
// @Failure 400 bad request
// @router /imagemanagement/v1/upload [post]
func (this *UploadController) Post() {
	log.Info("Upload post request received.")

	clientIp := this.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		this.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}

	this.displayReceivedMsg(clientIp)

	file, head, err := this.GetFile("file")
	if err != nil {
		this.HandleLoggingForError(clientIp, util.BadRequest, "Upload package file error")
		return
	}

	err = util.ValidateFileExtensionZip(head.Filename)
	if err != nil || len(head.Filename) > util.MaxFileNameSize {
		this.HandleLoggingForError(clientIp, util.BadRequest,
			"File shouldn't contains any extension or filename is larger than max size")
		return
	}

	err = util.ValidateFileSize(head.Size, util.MaxAppPackageFile)
	if err != nil {
		this.HandleLoggingForError(clientIp, util.BadRequest, "File size is larger than max size")
		return
	}
	defer file.Close()

	filename := head.Filename //original name for file

	//userId := this.Ctx.Input.Query("userId"), 加对userId字段的判断

	userId := this.GetString("userId")
	priority := this.GetString("priority")

	//创建imageId,fileName, uploadTime, userId
	imageId := createImageID()

	//change
	storageMedium := this.getStorageMedium(priority)

	err = createDirectory(storageMedium)
	if err != nil {
		log.Error("failed to create file path" + storageMedium)
	}

	saveFileName := imageId + filename //9c73996089944709bad8efa7f532aebe+1.zip
	err = this.SaveToFile("file", storageMedium+saveFileName)

	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to upload package")
		return
		//修改response code，加错误信息
	} else {
		//this.Ctx.WriteString("upload success")
		log.Info("save file to " + storageMedium)
	}

	//feedback download url to user
	url := "/imagemanagement/v1/download?imageId=" + imageId
	err = this.insertOrUpdateFileRecord(imageId, filename, userId, saveFileName, storageMedium, url)
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
		"url":           url})

	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return upload details")
		return
	}

	_, _ = this.Ctx.ResponseWriter.Write(uploadResp)

}
