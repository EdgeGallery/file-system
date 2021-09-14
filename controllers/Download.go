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

//压缩文件
//files 文件数组，可以是不同dir下的文件或者文件夹
//dest 压缩文件存放地址
func Compress(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, file := range files {
		err := compress(file, "", w)
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + info.Name() + "/"
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, zw)
			file.Close()
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + header.Name
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
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
// @router /image-management/v1/images/:imageId/action/download [get]
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
		filenameWithoutExt := strings.TrimSuffix(originalName, filepath.Ext(originalName))
		zipFilePath := filePath + filenameWithoutExt
		err := createDirectory(zipFilePath)
		if err != nil {
			log.Error("when compress, failed to create file path")
			return
		}
		_, err = CopyFile(downloadPath, zipFilePath+"/"+originalName)
		if err != nil {
			log.Error("when compress, failed to copy file")
			return
		}
		f1, err := os.Open(zipFilePath)
		if err != nil {
			log.Error("failed to open file")
			return
		}
		var files = []*os.File{f1}
		downloadName := strings.TrimSuffix(originalName, filepath.Ext(originalName)) + ".zip"
		err = Compress(files, filePath+downloadName)
		if err != nil {
			log.Error("failed to compress upload file")
			return
		}
		this.Ctx.Output.Download(downloadPath, downloadName)
		err = os.Remove(filePath+downloadName)
		if err != nil {
			this.writeErrorResponse(util.FailedToDeleteCache, util.StatusInternalServerError)
			return
		}
		err = os.RemoveAll(zipFilePath)
		if err != nil {
			this.writeErrorResponse(util.FailedToDeleteCache, util.StatusInternalServerError)
			return
		}
	} else {
		this.Ctx.Output.Download(downloadPath,originalName)
	}

}
