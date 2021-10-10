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
// @Description  Query and Delete api for filesystem
// @Author  GuoZhen Gao (2021/6/30 10:40)
package controllers

import (
	"encoding/json"
	"fileSystem/models"
	"fileSystem/util"
	log "github.com/sirupsen/logrus"
	"os"
)

// DownloadController   Define the Image controller to control query and delete
type ImageController struct {
	BaseController
}

// @Title Get
// @Description perform local image query operation
// @Param	imageId 	string
// @Success 200 ok
// @Failure 400 bad request
// @router /image-management/v1/images/:imageId [GET]
func (this *ImageController) Get() {
	log.Info("Query for local image get request received.")

	clientIp := this.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		this.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}

	this.displayReceivedMsg(clientIp)

	var imageFileDb models.ImageDB

	imageId := this.Ctx.Input.Param(":imageId")
	if imageId == "" {
		this.HandleLoggingForError(clientIp, util.StatusNotFound, "imageId is not right")
		return
	}

	_, err = this.Db.QueryTable("image_d_b", &imageFileDb, "image_id__exact", imageId)

	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusNotFound, "fail to query this imageId in database")
		return
	}

	filename := imageFileDb.SaveFileName
	uploadTime := imageFileDb.UploadTime.Format("2006-01-02 15:04:05")
	userId := imageFileDb.UserId
	storageMedium := imageFileDb.StorageMedium
	slimStatus := imageFileDb.SlimStatus

	var checkStatusResponse CheckStatusResponse
	var checkInfo CheckInfo
	var imageInfo ImageInfo

	if slimStatus == 2 {
		imageInfo.Filename = "compressed" + imageFileDb.SaveFileName
	} else {
		imageInfo.Filename = imageFileDb.SaveFileName
	}

	imageInfo.Format = imageFileDb.Format
	imageInfo.CheckErrors = imageFileDb.CheckErrors
	imageInfo.ImageEndOffset = imageFileDb.ImageEndOffset

	checkInfo.Checksum = imageFileDb.Checksum
	checkInfo.CheckResult = imageFileDb.CheckResult
	checkInfo.ImageInformation = imageInfo

	checkStatusResponse.Msg = imageFileDb.CheckMsg
	checkStatusResponse.Status = imageFileDb.CheckStatus
	checkStatusResponse.CheckInformation = checkInfo

	uploadResp, err := json.Marshal(map[string]interface{}{
		"imageId":             imageId,
		"fileName":            filename,
		"uploadTime":          uploadTime,
		"userId":              userId,
		"storageMedium":       storageMedium,
		"slimStatus":          slimStatus,
		"checkStatusResponse": checkStatusResponse,
	})

	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return query details")
		return
	}

	_, _ = this.Ctx.ResponseWriter.Write(uploadResp)

}

// @Title Delete
// @Description perform local image delete operation
// @Param	imageId 	string
// @Success 200 ok
// @Failure 400 bad request
// @router /image-management/v1/images/:imageId [DELETE]
func (this *ImageController) Delete() {
	log.Info("Delete local image package request received.")
	clientIp := this.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		this.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}

	this.displayReceivedMsg(clientIp)

	var imageFileDb models.ImageDB

	imageId := this.Ctx.Input.Param(":imageId")

	_, err = this.Db.QueryTable("image_d_b", &imageFileDb, "image_id__exact", imageId)

	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusNotFound, "fail to query this imageId in database")
		return
	}

	filename := imageFileDb.SaveFileName
	storageMedium := imageFileDb.StorageMedium

	file := storageMedium + filename

	err = os.Remove(file)

	fileRecord := &models.ImageDB{
		ImageId: imageId,
	}

	err = this.Db.DeleteData(fileRecord, "image_id")
	if err != nil && err.Error() != util.LastInsertIdNotSupported {
		this.HandleLoggingForError(clientIp, util.StatusInternalServerError, err.Error())
		return
	}

	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to delete package in vm")
		return
	} else {
		this.Ctx.WriteString("delete success")
		log.Info("delete file from " + storageMedium)
	}

}
