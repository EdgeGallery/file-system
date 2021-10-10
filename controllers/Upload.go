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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fileSystem/models"
	"fileSystem/util"
	"fmt"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadController   Define the controller to control upload
type UploadController struct {
	BaseController
}

type CheckResponse struct {
	Status    int    `json:"status"`
	Msg       string `json:"msg"`
	RequestId string `json:"requestId"`
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

func CreateImageID() string {
	uuId := uuid.NewV4()
	//return strings.Replace(uuId.String(), "-", "", -1)
	return uuId.String()
}

func (c *UploadController) insertOrUpdateFileRecord(imageId, fileName, userId, saveFileName, storageMedium, requestIdCheck string) error {

	fileRecord := &models.ImageDB{
		ImageId:        imageId,
		FileName:       fileName,
		UserId:         userId,
		SaveFileName:   saveFileName,
		StorageMedium:  storageMedium,
		RequestIdCheck: requestIdCheck,
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

// 编写一个函数，接收两个文件路径:
//srcFile := "e:/copyFileTest02.pdf" -- 源文件路径
//dstFile := "e:/Go/tools/copyFileTest02.pdf" -- 目标文件路径
func CopyFile(srcFileName string, dstFileName string) (written int64, err error) {
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

	clientIp, err, file, head, isDone := c.foreCheck()
	if isDone {
		return
	}

	defer file.Close()

	filename := head.Filename //original name for file   1.zip or 1.qcow2
	userId := c.GetString(util.UserId)
	priority := c.GetString(util.Priority)
	imageId := CreateImageID()

	storageMedium := c.getStorageMedium(priority)
	saveFileName := imageId + filename //9c73996089944709bad8efa7f532aebe+   1.zip or  1.qcow2
	err = c.saveByPriority(priority, saveFileName)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to upload package")
		return
	}
	originalName := filename

	//if file is zip file, decompress it to image file
	if filepath.Ext(head.Filename) == ".zip" {
		filenameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
		decompressFilePath := storageMedium + saveFileName
		arr, err := DeCompress(decompressFilePath, storageMedium)
		if err != nil {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, util.FailedToDecompress)
			return
		}
		originalName = subString(arr[0], strings.LastIndex(arr[0], "/")+1, len(arr[0]))
		saveFileName = imageId + originalName
		srcFileName := storageMedium + filenameWithoutExt + "/" + originalName
		dstFileName := storageMedium + saveFileName
		_, err = CopyFile(srcFileName, dstFileName)
		if err != nil {
			log.Error("when decompress, failed to copy file")
			return
		}
		err = os.Remove(decompressFilePath)
		if err != nil {
			c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to delete tmp zip file in vm")
			return
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	var formConfigMap map[string]string
	formConfigMap = make(map[string]string)
	formConfigMap["inputImageName"] = saveFileName

	requestJson, _ := json.Marshal(formConfigMap)
	requestBody := bytes.NewReader(requestJson)

	response, err := client.Post("http://localhost:5000/api/v1/vmimage/check", "application/json", requestBody)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "cannot send request to imagesOps")
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	var checkResponse CheckResponse
	err = json.Unmarshal(body, &checkResponse)
	if err != nil {
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
	}

	status := checkResponse.Status
	msg := checkResponse.Msg
	requestIdCheck := checkResponse.RequestId

	err = c.insertOrUpdateFileRecord(imageId, originalName, userId, saveFileName, storageMedium, requestIdCheck)
	if err != nil {
		log.Error("fail to insert imageID, filename, userID to database")
		return
	}

	uploadResp, err := json.Marshal(map[string]interface{}{
		"imageId":       imageId,
		"fileName":      originalName,
		"uploadTime":    time.Now().Format("2006-01-02 15:04:05"),
		"userId":        userId,
		"storageMedium": storageMedium,
		"isSlimmed":     0, //[0,1,2,3]  未瘦身/瘦身中/成功/失败
		"status":        status,
		"msg":           msg,
	})

	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return upload details")
		return
	}
	_, _ = c.Ctx.ResponseWriter.Write(uploadResp)
}

func (c *UploadController) foreCheck() (string, error, multipart.File, *multipart.FileHeader, bool) {
	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return "", nil, nil, nil, true
	}
	c.displayReceivedMsg(clientIp)

	file, head, err := c.GetFile("file")
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, "Upload package file error")
		return "", nil, nil, nil, true
	}

	err = util.ValidateFileExtension(head.Filename)
	if err != nil || len(head.Filename) > util.MaxFileNameSize {
		c.HandleLoggingForError(clientIp, util.BadRequest,
			"File shouldn't contains any extension or filename is larger than max size")
		return "", nil, nil, nil, true
	}

	err = util.ValidateFileSize(head.Size, util.MaxAppPackageFile)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, "File size is larger than max size")
		return "", nil, nil, nil, true
	}
	return clientIp, err, file, head, false
}
