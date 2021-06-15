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
	"archive/zip"
	"errors"
	"fileSystem/models"
	"fileSystem/util"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//下载文件
type DownloadController struct {
	BaseController
}

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

/*//解压
func DeCompress(zipFile, dest string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		filename := dest + file.Name
		err = os.MkdirAll(getDir(filename), 0755)
		if err != nil {
			return err
		}
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}
		w.Close()
		rc.Close()
	}
	return nil
}

func getDir(path string) string {
	return subString(path, 0, strings.LastIndex(path, "/"))
}

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
}*/

// extract zip package
func (this *DownloadController) extractZipPackage(packagePath string) (string, error) {
	zipReader, err := zip.OpenReader(packagePath)
	if err != nil {
		return "", errors.New("fail to open zip file")
	}
	if len(zipReader.File) != util.SingleFile {
		return "", errors.New("only support one image file in zip")
	}

	var totalWrote int64
	packageDir := path.Dir(packagePath) //destination path for file to save in linux
	err = os.MkdirAll(packageDir, 0750)
	if err != nil {
		log.Error(util.FailedToMakeDir)
		return "", errors.New(util.FailedToMakeDir)
	}
	for _, file := range zipReader.Reader.File {

		zippedFile, err := file.Open()
		if err != nil || zippedFile == nil {
			log.Error("Failed to open zip file")
			continue
		}
		if file.UncompressedSize64 > util.SingleFileTooBig || totalWrote > util.TooBig {
			log.Error("File size limit is exceeded")
		}

		defer zippedFile.Close()

		isContinue, wrote := this.extractFiles(file, zippedFile, totalWrote, packageDir)
		if isContinue {
			continue
		}
		totalWrote = wrote
	}
	return packageDir, nil
}

// Extract files
func (this *DownloadController) extractFiles(file *zip.File, zippedFile io.ReadCloser, totalWrote int64, dirName string) (bool, int64) {
	targetDir := dirName
	extractedFilePath := filepath.Join(
		targetDir,
		file.Name,
	)

	if file.FileInfo().IsDir() {
		err := os.MkdirAll(extractedFilePath, 0750)
		if err != nil {
			log.Error("Failed to create directory")
		}
	} else {
		outputFile, err := os.OpenFile(
			extractedFilePath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0750,
		)
		if err != nil || outputFile == nil {
			log.Error("The output file is nil")
			return true, totalWrote
		}

		defer outputFile.Close()

		wt, err := io.Copy(outputFile, zippedFile)
		if err != nil {
			log.Error("Failed to copy zipped file")
		}
		totalWrote += wt
	}
	return false, totalWrote
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

	if this.Ctx.Input.Query("iszip") == "true" {
		downloadName := strings.TrimSuffix(originalName,filepath.Ext(originalName)) + ".zip"
		this.Ctx.Output.Download(downloadPath, downloadName)
	} else {
		newPath, err := this.extractZipPackage(downloadPath)
		if err != nil {
			this.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailedToDecompress)
			return
		}
		downloadPath = newPath + "/" + imageId + originalName
		this.Ctx.Output.Download(downloadPath, originalName)
		err = os.Remove(downloadPath)
		if err != nil {
			this.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailedToDeleteCache)
			return
		}
	}

}
