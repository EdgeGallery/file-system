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
	"errors"
	"fileSystem/util"
	log "github.com/sirupsen/logrus"
	"os"
)

//ChunkUploadController  Define the controller to control upload
type UploadChunkController struct {
	BaseController
}

//add more storage logic here
func (c *UploadChunkController) getStorageMedium(priority string) string {
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
// @router "/image-management/v1/images/upload [get]
func (c *UploadChunkController) Get() {
	log.Info("Upload get request received.")
	c.Ctx.WriteString("Upload get request received.")
}

// @Title saveByPriority
// @Description upload file
// @Param   priority        string      true   "priority "
// @Param   saveFilename 	string  	true   "file"         eg.1.part
// @Param   identifier 	    string  	true   "file"         eg. xxxxxxxxxxxxx
func (c *UploadChunkController) saveByIdentifier(priority string, saveFilename string, identifier string) error {
	switch {
	case priority == "A":
		return errors.New("sorry, this storage medium is not supported right now")

	default:
		defaultPath := util.LocalStoragePath // "/usr/app/vmImage/"
		saveFilePath := defaultPath + identifier + "/"

		err := createDirectory(saveFilePath)
		if err != nil {
			log.Error("failed to create file path: " + saveFilePath)
			return err
		}

		err = c.SaveToFile(util.Part, saveFilePath+saveFilename)
		if err != nil {
			c.writeErrorResponse("fail to upload chunk file part", util.StatusInternalServerError)
			return err
		} else {
			log.Info("save file to " + saveFilePath)
			return nil
		}
	}
}

// @Title Post
// @Description upload file
// @Param   identifier  form-data 	string	true   "identifier"
// @Param   priority    form-data   string  true   "priority "
// @Param   file        form-data 	file	true   "file"
// @Success 200 ok
// @Failure 400 bad request
// @router /image-management/v1/images/upload [post]
func (c *UploadChunkController) Post() {
	log.Info("Upload post request received.")
	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}
	c.displayReceivedMsg(clientIp)

	chunkFile, head, err := c.GetFile("part")

	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, "Upload package file error")
		return
	}

	defer chunkFile.Close()

	identifier := c.GetString(util.Identifier)
	priority := c.GetString(util.Priority)

	//use identifier to create saving path
	err = c.saveByIdentifier(priority, head.Filename, identifier)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to save chunk file part")
		return
	}

	c.Ctx.WriteString("ok.")
}

// @Title Delete
// @Description perform local image delete operation
// @Param	identifier 	string
// @Success 200 ok
// @Failure 400 bad request
// @router /image-management/v1/images/upload [DELETE]
func (c *UploadChunkController) Delete() {
	log.Info("Delete local part file request received.")
	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}

	c.displayReceivedMsg(clientIp)

	priority := c.GetString(util.Priority)
	identifier := c.GetString(util.Identifier)

	storageMedium := c.getStorageMedium(priority)

	saveFilePath := storageMedium + identifier + "/"

	err = os.RemoveAll(saveFilePath)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to delete part file in vm")
		return
	} else {
		c.Ctx.WriteString("cancel success")
		log.Info("delete file from " + storageMedium + identifier + "/")
	}

}
