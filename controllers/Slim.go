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

type CompressResult struct {
	Status    int    `json:"status"`
	Msg       string `json:"msg"`
	RequestId string `json:"requestId"`
}

type ImageInfo struct {
	ImageEndOffset int    `json:"image-end-offset"`
	CheckErrors    int    `json:"check-errors"`
	Format         string `json:"format"`
	Filename       string `json:"filename"`
}

type CheckInfo struct {
	Checksum         string    `json:"checksum"`
	CheckResult      int       `json:"checkResult"`
	ImageInformation ImageInfo `json:"imageInfo"`
}

type CheckStatusResponse struct {
	Status           int       `json:"status"`
	Msg              string    `json:"msg"`
	CheckInformation CheckInfo `json:"checkInfo"`
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

// @Title Post
// @Description perform image slim operation
// @Param	imageId 	string
// @Success 200 ok
// @Failure 400 bad request
// @Success 200 ok
// @Failure 400 bad request
// @router "/image-management/v1/images/:imageId/action/slim [post]
func (c *SlimController) Post() {
	log.Info("image slim request received.")

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
	userId := imageFileDb.UserId
	storageMedium := imageFileDb.StorageMedium
	//originalFilename := imageFileDb.FileName

	if !c.PathCheck(filePath) {
		c.HandleLoggingForError(clientIp, util.StatusNotFound, "file path doesn't exist")
		return
	}

	//TODO: 添加isSlim字段后，先判断是否已经瘦身，若瘦身，则返回
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	/*v := url.Values{}
	v.Set("inputImageName", saveFilename)
	v.Set("outputImageName", "compressed"+saveFilename)*/
	var formConfigMap map[string]string
	formConfigMap = make(map[string]string)
	formConfigMap["inputImageName"] = saveFilename
	formConfigMap["outputImageName"] = "compressed" + saveFilename

	requestJson, _ := json.Marshal(formConfigMap)
	requestBody := bytes.NewReader(requestJson)

	response, err := client.Post("http://imageops/api/v1/vmimage/compress", "application/json", requestBody)
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

	//status := compressRes.Status
	/*
		加线程等待还是？
	*/
	/*	if status == 1 {
		}else if status == 0 {
		}*/
	requestId := compressRes.RequestId
	responseCheck, err := client.Get("http://imageops/api/v1/vmimage/" + requestId)
	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to request vmimage check")
		return
	}
	defer responseCheck.Body.Close()
	bodyCheck, err := ioutil.ReadAll(responseCheck.Body)

	var checkStatusResponse CheckStatusResponse
	err = json.Unmarshal(bodyCheck, &checkStatusResponse)

	if err != nil {
		c.writeErrorResponse(util.FailedToUnmarshal, util.BadRequest)
		return
	}

	checkResp, err := json.Marshal(map[string]interface{}{
		"imageId":       imageId,
		"uploadTime":    time.Now().Format("2006-01-02 15:04:05"),
		"userId":        userId,
		"storageMedium": storageMedium,
		"isSlim":      true,
		"checksum":    checkStatusResponse.CheckInformation.Checksum,
		"imageInfo":  checkStatusResponse.CheckInformation,
	})

	if err != nil {
		c.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return check details")
		return
	}
	_, _ = c.Ctx.ResponseWriter.Write(checkResp)
}
