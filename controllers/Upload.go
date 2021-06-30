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
// @Description  Upload api for filesystem
// @Author  GuoZhen Gao (2021/6/30 10:40)
package controllers

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"fileSystem/models"
	"fileSystem/util"
	"fmt"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadController   Define the controller to control upload
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

// @Title saveByPriority
// @Description upload file
// @Param   priority     string  true   "priority "
// @Param   saveFilename 	string  	true   "file"   eg.9c73996089944709bad8efa7f532aebe1.zip
func (c *UploadController) saveByPriority(priority string, saveFilename string) error {
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

// 编写一个函数，接收两个文件路径:
//srcFile := "e:/copyFileTest02.pdf" -- 源文件路径
//dstFile := "e:/Go/tools/copyFileTest02.pdf" -- 目标文件路径
func copyFile(srcFileName string, dstFileName string) (written int64, err error) {
	srcFile, err := os.Open(srcFileName)
	if err != nil {
		fmt.Printf("open file error = %v\n", err)
	}
	defer srcFile.Close()

	//通过srcFile，获取到READER
	reader := bufio.NewReader(srcFile)

	//打开dstFileName
	dstFile, err := os.OpenFile(dstFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("open file error = %v\n", err)
		return
	}

	//通过dstFile，获取到WRITER
	writer := bufio.NewWriter(dstFile)
	//writer.Flush()

	defer dstFile.Close()

	return io.Copy(writer, reader)
}

// @Title Get
// @Description test connection is ok or not
// @Success 200 ok
// @Failure 400 bad request
// @router "/image-management/v1/images [get]
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
// @router "/image-management/v1/images [post]
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
	err = util.ValidateFileExtension(head.Filename)
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

	filename := head.Filename //original name for file   1.zip or 1.qcow2
	userId := c.GetString(util.UserId)
	priority := c.GetString(util.Priority)

	//create imageId, fileName, uploadTime, userId
	imageId := createImageID()

	//get a storage medium to let fe know
	storageMedium := c.getStorageMedium(priority)
	saveFileName := imageId + filename //9c73996089944709bad8efa7f532aebe+   1.zip or  1.qcow2
	err = c.saveByPriority(priority, saveFileName)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to upload package")
		return
	}

	//if file is not zip file, compress it to zip
	if filepath.Ext(head.Filename) != ".zip" {
		originalName := strings.TrimSuffix(filename, filepath.Ext(filename))
		zipFilePath := storageMedium + originalName
		err := createDirectory(zipFilePath)
		if err != nil {
			log.Error("when compress, failed to create file path to" + zipFilePath)
			return
		}
		_, err = copyFile(storageMedium+saveFileName, zipFilePath+"/"+filename)
		if err != nil {
			log.Error("when compress, failed to create file path to" + zipFilePath)
			return
		}
		f1, err := os.Open(zipFilePath)
		if err != nil {
			log.Error("failed to open upload file")
			return
		}
		var files = []*os.File{f1}
		newSaveFileName := strings.TrimSuffix(saveFileName, filepath.Ext(filename)) //9c73996089944709bad8efa7f532aebe+1
		err = Compress(files, storageMedium+newSaveFileName+".zip")
		if err != nil {
			log.Error("failed to compress upload file")
			return
		}
		err = os.Remove(storageMedium + saveFileName)
		if err != nil {
			c.writeErrorResponse(util.FailedToDeleteCache, util.StatusInternalServerError)
			return
		}
		err = os.RemoveAll(zipFilePath)
		if err != nil {
			c.writeErrorResponse(util.FailedToDeleteCache, util.StatusInternalServerError)
			return
		}
		saveFileName = newSaveFileName + ".zip"
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
