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
	"fileSystem/models"
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
	BaseUrl       string = "http://edgegallery:9500/image-management/v1"
	imageId       string = "94d6e70d-51f7-4b0d-965f-59dca2c3002c"
	UserIdKey     string = "userId"
	UserId        string = "71ea862b-5806-4196-bce3-434bf9c95b18"
	requestId     string = "71ea862b-5806-4196-bce3-434bf9c95b18"
	PriorityKey   string = "priority"
	Priority      string = "0"
	originalName         = "cirros.qcow2"
	storageMedium        = "/usr/app/vmImage/"
	saveFileName         = "71ea862b-5806-4196-bce3-434bf9c95b18cirros.qcow2"
	err                  = errors.New("error")

	Post   = "POST"
	Put    = "PUT"
	Get    = "GET"
	Delete = "DELETE"
)

func TestControllerSuccess(t *testing.T) {

	// Common steps
	// Setting file path
	// return filesystem/test的目录地址
	path, _ := os.Getwd()
	path += "/mockImage.qcow2"

	// Setting extra parameters
	extraParams := map[string]string{
		UserIdKey:   UserId,
		PriorityKey: Priority,
	}

	testDb := &MockDb{
		imageRecords: make(map[string]models.ImageDB),
	}

	var c *beego.Controller
	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(c), "ServeJSON", func(*beego.Controller, ...bool) {
		go func() {
			// do nothing
		}()
	})
	defer patch1.Reset()

	//testUploadGet(t, extraParams, "", testDb)

	testUploadPostValidateSrcAddressErr(t, extraParams, path, testDb)
	testUploadPostValidateSrcAddress(t, extraParams, path, testDb)

	testUploadPostImageOpsPostGetOk(t, extraParams, path, testDb)
}

/*func testUploadGet(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {

	t.Run("testUploadGet", func(t *testing.T) {

		//GET Request
		queryRequest, _ := getHttpRequest("http://edgegallery:9500/image-management/v1/images",
			extraParams, "file", path, "GET", []byte(""))

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

		// Test query
		queryController.Get()

		// Check for success case wherein the status value will be default i.e. 0
		assert.Equal(t, 0, queryController.Ctx.ResponseWriter.Status, "Upload get request received.")
		_ = queryController.Ctx.ResponseWriter.ResponseWriter.(*httptest.ResponseRecorder)
	})
}*/

func testUploadPostValidateSrcAddressErr(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {

	t.Run("testUploadPost", func(t *testing.T) {
		//GET Request
		queryRequest, _ := getHttpRequest("http://edgegallery:9500/image-management/v1/images",
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

		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return err
		})
		defer patch1.Reset()

		// Test query
		queryController.Post()

	})
}

func testUploadPostValidateSrcAddress(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {

	t.Run("testUploadPost", func(t *testing.T) {
		//GET Request
		queryRequest, _ := getHttpRequest("http://edgegallery:9500/image-management/v1/images",
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

		//GET Request
		queryRequest, _ := getHttpRequest("http://edgegallery:9500/image-management/v1/images",
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

/*func TestCronGetCheck(t *testing.T) {

	getBeegoController := beego.Controller{Ctx: &context.Context{ResponseWriter: &context.Response{ResponseWriter: httptest.NewRecorder()}},
		Data: make(map[interface{}]interface{})}

	testDb := &MockDb{
		imageRecords: make(map[string]models.ImageDB),
	}
	uploadController := &controllers.UploadController{BaseController: controllers.BaseController{Db: testDb,
		Controller: getBeegoController}}

	var responseGetBodyMap map[string]interface{}
	responseGetBodyMap = make(map[string]interface{})

	var imageInfo controllers.ImageInfo
	var checkInfo controllers.CheckInfo
	var checkStatusResponse controllers.CheckStatusResponse

	checkStatusResponse.Status = 6
	checkStatusResponse.Msg = "Check Time Out"

	responseGetBodyMap["status"] = 6
	responseGetBodyMap["msg"] = "Check Time Out"

	imageInfo.ImageEndOffset = "564330496"
	imageInfo.CheckErrors = "0"
	imageInfo.Format = "qcow2"
	imageInfo.Filename = "ubuntu-18.04.qcow2"
	checkInfo.ImageInformation = imageInfo
	checkInfo.CheckResult = 0
	checkInfo.Checksum = "782fa5257615748e673eefe0143188e4"
	checkStatusResponse.CheckInformation = checkInfo
	responseGetBodyMap["checkInfo"] = checkInfo

	responseGetJson, _ := json.Marshal(responseGetBodyMap)
	responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

	patch3 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
		return &http.Response{Body: responseGetBody}, nil
	})
	defer patch3.Reset()

	uploadController.CronGetCheck(requestId, imageId, "name", UserId, "/usr/app/vmImage/", "saveFileName")

}*/

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
	checkResponse,_ := c.PostToCheck("SaveFileName")
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
