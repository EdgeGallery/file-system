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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fileSystem/controllers"
	"fileSystem/models"
	"fileSystem/pkg/dbAdpater"
	"fileSystem/util"
	"github.com/agiledragon/gomonkey"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestSlimController(t *testing.T) {
	// Common steps
	// Setting file path
	// return filesystem/test的目录地址
	path, _ := os.Getwd()

	// Setting extra parameters
	extraParams := map[string]string{
		UserIdKey:   UserId,
		PriorityKey: Priority,
	}

	fileRecordSlimmed := models.ImageDB{
		ImageId:       imageId,
		FileName:      util.FileName,
		UserId:        UserId,
		SaveFileName:  saveFileName,
		StorageMedium: storageMedium,
		SlimStatus:    2,
	}
	testDb := &MockDb{
		imageRecords: fileRecordSlimmed,
	}

	var c *beego.Controller
	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(c), "ServeJSON", func(*beego.Controller, ...bool) {
		go func() {
			// do nothing
		}()
	})
	defer patch1.Reset()

	slimController := getSlimController(extraParams, path, testDb)
	testSlimIpErr(slimController, t)
	testSlimCompressPostErr(slimController, t)
	testSlimInProgress(slimController, t)
	testSlimCompressFailed(slimController, t)
	testSlimCompressElse(slimController, t)
	testAsyCallImageOpsGetCompressErr(slimController, t, fileRecordSlimmed)
	testAsyCallCompressCompleted(slimController, t, fileRecordSlimmed)
	testAsyCallCompressInProgress(slimController, t, fileRecordSlimmed)
	testAsyCallCompressFailed(slimController, t, fileRecordSlimmed)
	testAsyCallCompressNoEnoughSpace(slimController, t, fileRecordSlimmed)
	testAsyCallCompressTimeout(slimController, t, fileRecordSlimmed)
	testAsyCallCompressInsertError(slimController, t, fileRecordSlimmed)
	testAsyCallImageOps(slimController, t, fileRecordSlimmed)

}

func testSlimIpErr(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testSlimIpErr", func(t *testing.T) {
		// Test query
		slimController.Post()
	})
}

func testSlimCompressPostErr(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testSlimCompressPostErr", func(t *testing.T) {
		// Test query
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "PathCheck",
			func(_ *controllers.SlimController, _ string) bool {
				return true
			})
		defer patch2.Reset()
		slimController.Post()
	})
}

func testSlimInProgress(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testSlimInProgress", func(t *testing.T) {
		// Test query
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "PathCheck",
			func(_ *controllers.SlimController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		var responsePostBodyMap map[string]interface{}
		responsePostBodyMap = make(map[string]interface{})
		responsePostBodyMap["status"] = 0
		responsePostBodyMap["msg"] = "Compress In Progress"
		responsePostBodyMap["requestId"] = requestId
		responsePostJson, _ := json.Marshal(responsePostBodyMap)
		responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client,
			url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return &http.Response{Body: responsePostBody}, nil
		})
		defer patch3.Reset()

		slimController.Post()
	})
}

func testSlimCompressFailed(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testSlimCompressFailed", func(t *testing.T) {
		// Test query
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "PathCheck",
			func(_ *controllers.SlimController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		var responsePostBodyMap map[string]interface{}
		responsePostBodyMap = make(map[string]interface{})
		responsePostBodyMap["status"] = 1
		responsePostBodyMap["msg"] = "Compress Failed"
		responsePostBodyMap["requestId"] = requestId
		responsePostJson, _ := json.Marshal(responsePostBodyMap)
		responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client,
			url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return &http.Response{Body: responsePostBody}, nil
		})
		defer patch3.Reset()

		slimController.Post()
	})
}

func testSlimCompressElse(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testSlimCompressElse", func(t *testing.T) {
		// Test query
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "PathCheck",
			func(_ *controllers.SlimController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		var responsePostBodyMap map[string]interface{}
		responsePostBodyMap = make(map[string]interface{})
		responsePostBodyMap["status"] = 3
		responsePostBodyMap["msg"] = "Compress Internal error"
		responsePostBodyMap["requestId"] = requestId
		responsePostJson, _ := json.Marshal(responsePostBodyMap)
		responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client,
			url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return &http.Response{Body: responsePostBody}, nil
		})
		defer patch3.Reset()

		slimController.Post()
	})
}

func testAsyCallImageOpsGetCompressErr(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallImageOpsGetCompressErr", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		slimController.AsyCallImageOps(client, requestId, "127.0.0.1", imageFileDb, imageId)
	})
}

func testAsyCallCompressCompleted(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressCompleted", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		var responseGetBodyMap map[string]interface{}
		responseGetBodyMap = make(map[string]interface{})
		responseGetBodyMap["status"] = 0
		responseGetBodyMap["msg"] = "compress completed"
		responseGetBodyMap["rate"] = 1
		responseGetJson, _ := json.Marshal(responseGetBodyMap)
		responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		var responsePostBodyMap map[string]interface{}
		responsePostBodyMap = make(map[string]interface{})
		responsePostBodyMap["status"] = 0
		responsePostBodyMap["msg"] = "Compress In Progress"
		responsePostBodyMap["requestId"] = requestId
		responsePostJson, _ := json.Marshal(responsePostBodyMap)
		responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client,
			url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return &http.Response{Body: responsePostBody}, nil
		})
		defer patch3.Reset()

		slimController.AsyCallImageOps(client, requestId, "127.0.0.1", imageFileDb, imageId)
	})
}

func testAsyCallCompressInProgress(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressInProgress", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		var responseGetBodyMap map[string]interface{}
		responseGetBodyMap = make(map[string]interface{})
		responseGetBodyMap["status"] = 1
		responseGetBodyMap["msg"] = "compress in progress"
		responseGetBodyMap["rate"] = 0.5
		responseGetJson, _ := json.Marshal(responseGetBodyMap)
		responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCompressGetRecord",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ string, _ int,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, "127.0.0.1", imageFileDb, imageId)
	})
}

func testAsyCallCompressFailed(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressFailed", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		var responseGetBodyMap map[string]interface{}
		responseGetBodyMap = make(map[string]interface{})
		responseGetBodyMap["status"] = 2
		responseGetBodyMap["msg"] = "compress failed"
		responseGetBodyMap["rate"] = 0
		responseGetJson, _ := json.Marshal(responseGetBodyMap)
		responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCompressGetRecord",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ string, _ int,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, "127.0.0.1", imageFileDb, imageId)
	})
}

func testAsyCallCompressNoEnoughSpace(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressNoEnoughSpace", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		var responseGetBodyMap map[string]interface{}
		responseGetBodyMap = make(map[string]interface{})
		responseGetBodyMap["status"] = 3
		responseGetBodyMap["msg"] = "compress NoEnoughSpace"
		responseGetBodyMap["rate"] = 0
		responseGetJson, _ := json.Marshal(responseGetBodyMap)
		responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCompressGetRecord",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ string, _ int,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, "127.0.0.1", imageFileDb, imageId)
	})
}

func testAsyCallCompressTimeout(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressTimeout", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		var responseGetBodyMap map[string]interface{}
		responseGetBodyMap = make(map[string]interface{})
		responseGetBodyMap["status"] = 4
		responseGetBodyMap["msg"] = "compress Timeout"
		responseGetBodyMap["rate"] = 0
		responseGetJson, _ := json.Marshal(responseGetBodyMap)
		responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCompressGetRecord",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ string, _ int,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, "127.0.0.1", imageFileDb, imageId)
	})
}

func testAsyCallCompressInsertError(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressInsertError", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		var responseGetBodyMap map[string]interface{}
		responseGetBodyMap = make(map[string]interface{})
		responseGetBodyMap["status"] = 0
		responseGetBodyMap["msg"] = "compress success"
		responseGetBodyMap["rate"] = 1
		responseGetJson, _ := json.Marshal(responseGetBodyMap)
		responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCompressGetRecord",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ string, _ int,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, "127.0.0.1", imageFileDb, imageId)
	})
}

func testAsyCallImageOps(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallImageOps", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		var responseGetBodyMap map[string]interface{}
		responseGetBodyMap = make(map[string]interface{})
		responseGetBodyMap["status"] = 0
		responseGetBodyMap["msg"] = "compress completed"
		responseGetBodyMap["rate"] = 1
		responseGetJson, _ := json.Marshal(responseGetBodyMap)
		responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))

		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		var responsePostBodyMap map[string]interface{}
		responsePostBodyMap = make(map[string]interface{})
		responsePostBodyMap["status"] = 0
		responsePostBodyMap["msg"] = "Compress In Progress"
		responsePostBodyMap["requestId"] = requestId
		responsePostJson, _ := json.Marshal(responsePostBodyMap)
		responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client,
			url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return &http.Response{Body: responsePostBody}, nil
		})
		defer patch3.Reset()

		slimController.AsyCallImageOps(client, requestId, "127.0.0.1", imageFileDb, imageId)
	})
}

func getSlimController(extraParams map[string]string, path string, testDb dbAdpater.Database) *controllers.SlimController {
	//GET Request
	queryRequest, _ := getHttpRequest(UploadUrl+ZipUri,
		extraParams, "file", path, "GET", []byte(""))

	// Prepare Input
	queryInput := &context.BeegoInput{Context: &context.Context{Request: queryRequest}}
	setParam(queryInput, true)

	// Prepare beego controller
	queryBeegoController := beego.Controller{Ctx: &context.Context{Input: queryInput, Request: queryRequest,
		ResponseWriter: &context.Response{ResponseWriter: httptest.NewRecorder()}},
		Data: make(map[interface{}]interface{})}

	// Create Upload controller with mocked DB and prepared Beego controller
	queryController := &controllers.SlimController{controllers.BaseController{Db: testDb,
		Controller: queryBeegoController}}
	return queryController
}
