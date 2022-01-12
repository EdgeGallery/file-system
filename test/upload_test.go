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

package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fileSystem/controllers"
	"fileSystem/pkg/dbAdpater"
	"fileSystem/util"
	"github.com/agiledragon/gomonkey"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

var (
	imageId       string = "94d6e70d-51f7-4b0d-965f-59dca2c3002c"
	UserIdKey     string = "userId"
	UserId        string = "71ea862b-5806-4196-bce3-434bf9c95b18"
	requestId     string = "71ea862b-5806-4196-bce3-434bf9c95b18"
	PriorityKey   string = "priority"
	Priority      string = "0"
	storageMedium        = "/usr/app/vmImage/"
	saveFileName         = "71ea862b-5806-4196-bce3-434bf9c95b18cirros.qcow2"
	err                  = errors.New("error")
	UploadUrl            = "http://edgegallery:9500/image-management/v1/images"
	ZipUri               = "/94d6e70d-51f7-4b0d-965f-59dca2c3002c/action/download/?isZip=true"
	NotZipUri            = "/94d6e70d-51f7-4b0d-965f-59dca2c3002c/action/download"
)

func TestControllerSuccess(t *testing.T) {
	path, extraParams, testDb := prepareTest()
	var c *beego.Controller
	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(c), "ServeJSON", func(*beego.Controller, ...bool) {
		go func() {
			// do nothing
		}()
	})
	defer patch1.Reset()
	path += "/mockImage.qcow2"
	testUploadPostValidateSrcAddressErr(t, extraParams, path, testDb)
	testUploadPostValidateSrcAddress(t, extraParams, path, testDb)
	testUploadPostImageOpsPostGetOk(t, extraParams, path, testDb)
}

func testUploadPostValidateSrcAddressErr(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {

	t.Run("testUploadPost", func(t *testing.T) {
		//GET Request
		queryController := getQueryController(extraParams, path, testDb)

		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return err
		})
		defer patch1.Reset()

		// Test query
		queryController.Post()

	})
}

func getQueryController(extraParams map[string]string, path string, testDb dbAdpater.Database) *controllers.UploadController {
	queryRequest, _ := getHttpRequest(UploadUrl,
		extraParams, "file", path, "POST", []byte(""))

	// Prepare Input
	queryInput := &context.BeegoInput{Context: &context.Context{Request: queryRequest}}
	setParam(queryInput, false)

	// Prepare beego controller
	queryBeegoController := beego.Controller{Ctx: &context.Context{Input: queryInput, Request: queryRequest,
		ResponseWriter: &context.Response{ResponseWriter: httptest.NewRecorder()}},
		Data: make(map[interface{}]interface{})}

	// Create Upload controller with mocked DB and prepared Beego controller
	queryController := &controllers.UploadController{controllers.BaseController{Db: testDb,
		Controller: queryBeegoController}}
	return queryController
}

func testUploadPostValidateSrcAddress(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {

	t.Run("testUploadPost", func(t *testing.T) {
		queryController := getQueryController(extraParams, path, testDb)
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		// Test query
		queryController.Post()

	})
}

func testUploadPostImageOpsPostGetOk(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {

	t.Run("testUploadPost", func(t *testing.T) {

		queryController := getQueryController(extraParams, path, testDb)
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		var responsePostBodyMap map[string]interface{}
		responsePostBodyMap = make(map[string]interface{})
		responsePostBodyMap["status"] = 0
		responsePostBodyMap["msg"] = "Check In Progress"
		responsePostBodyMap["requestId"] = requestId
		responsePostJson, _ := json.Marshal(responsePostBodyMap)
		responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client, url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return &http.Response{Body: responsePostBody}, nil
		})
		defer patch2.Reset()

		// Test query
		queryController.Post()
	})
}

func TestUploadGetCheckInProgress(t *testing.T) {
	var responseBodyMapOfGet map[string]interface{}
	var imageInfo controllers.ImageInfo
	var checkInfo controllers.CheckInfo
	responseBodyMapOfGet = make(map[string]interface{})
	responseBodyMapOfGet["status"] = 0
	responseBodyMapOfGet["msg"] = "Check completed, the image is (now) consistent"
	imageInfo.ImageEndOffset = "564330496"
	imageInfo.CheckErrors = "0"
	imageInfo.Format = "qcow2"
	imageInfo.Filename = "ubuntu-18.04.qcow2"
	imageInfo.VirtualSize = 40.0
	imageInfo.DiskSize = "578359296"
	checkInfo.ImageInformation = imageInfo
	checkInfo.CheckResult = 0
	checkInfo.Checksum = "782fa5257615748e673eefe0143188e4"
	responseBodyMapOfGet["checkInfo"] = checkInfo

	responseGetJson, _ := json.Marshal(responseBodyMapOfGet)
	responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
		return &http.Response{Body: responseGetBody}, nil
	})
	defer patch1.Reset()

	c := getUploadController()
	c.GetToCheck(requestId)
}

func TestUploadPostToCheck(t *testing.T) {

	var responsePostBodyMap map[string]interface{}
	responsePostBodyMap = make(map[string]interface{})
	responsePostBodyMap["status"] = 0
	responsePostBodyMap["msg"] = "Check In Progress"
	responsePostBodyMap["requestId"] = requestId
	responsePostJson, _ := json.Marshal(responsePostBodyMap)
	responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client, url, contentType string, body io.Reader) (resp *http.Response, err error) {
		return &http.Response{Body: responsePostBody}, nil
	})

	defer patch1.Reset()

	c := getUploadController()
	checkResponse, _ := c.PostToCheck("SaveFileName")
	assert.Equal(t, 0, checkResponse.Status, "Post to Check is ok")

}

func getUploadController() *controllers.UploadController {
	c := &controllers.UploadController{}
	c.Init(context.NewContext(), "", "", nil)
	req, err := http.NewRequest("POST", "http://127.0.0.1", strings.NewReader(""))
	if err != nil {
		log.Error("Prepare http request failed")
	}
	c.Ctx.Request = req
	c.Ctx.Request.Header.Set("X-Real-Ip", "127.0.0.1")
	c.Ctx.ResponseWriter = &context.Response{}
	c.Ctx.ResponseWriter.ResponseWriter = httptest.NewRecorder()
	c.Ctx.Output = context.NewOutput()
	c.Ctx.Input = context.NewInput()
	c.Ctx.Output.Reset(c.Ctx)
	c.Ctx.Input.Reset(c.Ctx)
	c.Ctx.Input.SetParam(UserIdKey, UserId)
	c.Ctx.Input.SetParam(PriorityKey, Priority)
	path, _ := os.Getwd()
	c.Ctx.Input.SetParam("file", path+"/mockImage.qcow2")
	return c
}

func setParam(ctx *context.BeegoInput, isZip bool) {
	if isZip {
		ctx.SetParam("isZip", "true")
	}
	ctx.SetParam(":imageId", imageId)
}
