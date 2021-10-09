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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fileSystem/models"
	"fileSystem/util"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type SlimController struct {
	BaseController
}

// /vmimage/check post response
// /vmimage/compress post response
type CompressResult struct {
	Status    int    `json:"status"`
	Msg       string `json:"msg"`
	RequestId string `json:"requestId"`
}

type ImageInfo struct {
	ImageEndOffset string `json:"image-end-offset"`
	CheckErrors    string `json:"check-errors"`
	Format         string `json:"format"`
	Filename       string `json:"filename"`
}

type CheckInfo struct {
	Checksum         string    `json:"checksum"`
	CheckResult      int       `json:"checkResult"`
	ImageInformation ImageInfo `json:"imageInfo"`
}

// /vmimage/check/requestId get response
type CheckStatusResponse struct {
	Status           int       `json:"status"`
	Msg              string    `json:"msg"`
	CheckInformation CheckInfo `json:"checkInfo"`
}

// /vmimage/compress/requestId get response
type CompressStatusResponse struct {
	Status int     `json:"status"`
	Msg    string  `json:"msg"`
	Rate   float64 `json:"rate"`
}

// @Title PathCheck
// @Description check file in path is existed or not
// @Param   Source Zip File Path    string
func (c *SlimController) PathCheck(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func (c *SlimController) insertOrUpdatePostRecord(imageId string, slimStatus int, requestId string) error {

	fileRecord := &models.ImageDB{
		ImageId:           imageId,
		SlimStatus:        slimStatus,
		RequestIdCompress: requestId,
	}

	err := c.Db.InsertOrUpdateData(fileRecord, "image_id")

	if err != nil && err.Error() != "LastInsertId is not supported by this driver" {
		log.Error("Failed to save file record to database.")
		return err
	}

	log.Info("Add file record: %+v", fileRecord)
	return nil
}

// @Title Post
// @Description perform image slim operation
// @Param	imageId 	string
// @Success 200 ok
// @Failure 400 bad request
// @router "/image-management/v1/images/:imageId/action/slim [post]
func (c *SlimController) Post() {
	log.Info("image slim Post request received.")

	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}

	c.displayReceivedMsg(clientIp)

	var imageFileDb models.ImageDB

	imageId := c.Ctx.Input.Param(":imageId")

	_, err = c.Db.QueryTable("image_d_b", &imageFileDb, "image_id__exact", imageId)

	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "fail to query database")
		return
	}

	filePath := imageFileDb.StorageMedium
	saveFilename := imageFileDb.SaveFileName
	if !c.PathCheck(filePath) {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "file path doesn't exist")
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	var formConfigMap map[string]string
	formConfigMap = make(map[string]string)
	formConfigMap["inputImageName"] = saveFilename
	formConfigMap["outputImageName"] = "compressed" + saveFilename

	requestJson, _ := json.Marshal(formConfigMap)
	requestBody := bytes.NewReader(requestJson)

	//imageops/api/v1/vmimage/compress
	response, err := client.Post("http://localhost:5000/api/v1/vmimage/compress", "application/json", requestBody)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "cannot send request to imagesOps")
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	var compressRes CompressResult
	err = json.Unmarshal(body, &compressRes)
	if err != nil {
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
	}

	requestId := compressRes.RequestId
	responseStatus := compressRes.Status //0:compress in progress  1: compress failed
	var slimStatus int                   //[0,1,2,3]  未瘦身/瘦身中/成功/失败
	if responseStatus == 0 {
		slimStatus = 1
	} else if responseStatus == 1 {
		slimStatus = 3
	}
	err = c.insertOrUpdatePostRecord(imageId, slimStatus, requestId)
	if err != nil {
		log.Error("fail to insert imageId,slimStatus,requestId to database")
		return
	}
	c.Ctx.WriteString("true")
}

// @Title Get
// @Description perform image slim operation
// @Param	imageId 	string
// @Success 200 ok
// @Failure 400 bad request
// @router "/image-management/v1/images/:imageId/action/slim [get]
func (c *SlimController) Get() {
	log.Info("image slim Get request received.")
	clientIp := c.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}
	c.displayReceivedMsg(clientIp)

	var imageFileDb models.ImageDB
	imageId := c.Ctx.Input.Param(":imageId")
	_, err = c.Db.QueryTable("image_d_b", &imageFileDb, "image_id__exact", imageId)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "fail to query database")
		return
	}

	requestId := imageFileDb.RequestIdCompress
	if len(requestId) == 0 {
		c.Ctx.WriteString("requestId is empty, this image doesn't begin slimming yet")
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	//http://imageops/api/v1/vmimage/compress
	response, err := client.Get("http://localhost:5000/api/v1/vmimage/compress/" + requestId)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to request vmimage compress check")
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	log.Error("string:" + string(body))
	var compressStatusResponse CompressStatusResponse
	err = json.Unmarshal(body, &compressStatusResponse)
	if err != nil {
		log.Error("fail to request http://localhost:5000/api/v1/vmimage/compress/" + requestId)
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
		return
	}

	// 0: compress completed; 1: compress in progress; 2: compress failed
	compressStatus := compressStatusResponse.Status
	var slimStatus int //[0,1,2,3]  未瘦身/瘦身中/成功/失败
	if compressStatus == 0 {
		slimStatus = 2
	} else if compressStatus == 1 {
		slimStatus = 1
	} else if compressStatus == 2 {
		slimStatus = 3
	}

	err = c.insertOrUpdatePostRecord(imageId, slimStatus, requestId) //update slimStatus
	if err != nil {
		log.Error("fail to insert imageId,slimStatus,requestId to database")
		return
	}
	slimMsg := compressStatusResponse.Msg
	slimRate := compressStatusResponse.Rate

	checkResp, err := json.Marshal(map[string]interface{}{
		"imageId":    imageId,
		"uploadTime": time.Now().Format("2006-01-02 15:04:05"),
		"slimStatus": slimStatus,
		"slimMsg":    slimMsg,
		"slimRate":   slimRate,
	})

	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return check details")
		return
	}
	_, _ = c.Ctx.ResponseWriter.Write(checkResp)

}
