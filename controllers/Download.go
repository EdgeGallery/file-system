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
// @Description  download api for filesystem
// @Author  GuoZhen Gao (2021/6/30 10:40)
package controllers

import (
	"archive/zip"
	"fileSystem/models"
	"fileSystem/util"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DownloadController   Define the download controller
type DownloadController struct {
	BaseController
}

// @Title PathCheck
// @Description check file in path is existed or not
// @Param   Source Zip File Path    string
func (this *DownloadController) PathCheck(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// @Title DeCompress
// @Description Decompress file
// @Param   Source Zip File Path    string
// @Param   Destination File Path    string
func DeCompress(zipFile, dest string) ([]string, error) {
	var res []string
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		filename := dest + "/" + file.Name
		err = os.MkdirAll(getDir(filename), 0755)
		if err != nil {
			return nil, err
		}

		if len(filename)-1 == strings.LastIndex(filename, "/") {
			continue
		}
		w, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		defer w.Close()
		_, err = io.Copy(w, rc)
		res = append(res, filename)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

//Helper function to get file path
func getDir(path string) string {
	return subString(path, 0, strings.LastIndex(path, "/"))
}

//Helper function to get substring
func subString(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < start || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}

// @Title Get
// @Description Download file
// @Param   imageId        path 	string	true   "imageId"
// @Success 200 ok
// @Failure 400 bad request
// @router /imagemanagement/v1/download [get]
func (this *DownloadController) Get() {
	log.Info("Download get request received.")

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

	//err = this.Db.QueryForDownload("image_d_b", &imageFileDb, imageId) //表名
	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusNotFound, "fail to query database")
		return
	}

	filePath := imageFileDb.StorageMedium
	if !this.PathCheck(filePath) {
		this.HandleLoggingForError(clientIp, util.StatusNotFound, "file path doesn't exist")
		return
	}

	fileName := imageFileDb.SaveFileName
	originalName := imageFileDb.FileName

	downloadPath := filePath + fileName

	if this.Ctx.Input.Query("isZip") == "true" {
		downloadName := strings.TrimSuffix(originalName, filepath.Ext(originalName)) + ".zip"
		this.Ctx.Output.Download(downloadPath, downloadName)
	} else {
		saveName := strings.TrimSuffix(originalName, filepath.Ext(originalName))
		arr, err := DeCompress(downloadPath, filePath+saveName)
		if err != nil {
			this.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailedToDecompress)
			return
		}

		downloadPath = arr[0]
		originalName = subString(downloadPath, strings.LastIndex(downloadPath, "/")+1, len(downloadPath))
		this.Ctx.Output.Download(downloadPath, originalName)

		err = os.RemoveAll(filePath + saveName)
		if err != nil {
			this.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailedToDeleteCache)
			return
		}
	}

}
