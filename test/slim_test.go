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
	"reflect"
	"testing"
)

var (
	compressInProgress = "Compress In Progress"
	compressCompleted  = "compress completed"
	LocalIp            = "127.0.0.1"
)

func TestSlimController(t *testing.T) {
	fileRecordSlimmed := models.ImageDB{
		ImageId:       imageId,
		FileName:      util.FileName,
		UserId:        UserId,
		SaveFileName:  saveFileName,
		StorageMedium: storageMedium,
		SlimStatus:    2,
	}
	path, extraParams, testDb := prepareTest()
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
	testAsyCallCompressElse(slimController, t, fileRecordSlimmed)
	testAsyCallCompressInsertError(slimController, t, fileRecordSlimmed)
	testAsyCallImageOpsGetCheckErr(slimController, t, fileRecordSlimmed)
	testCheckResponseInProgress(slimController, t)
	testCheckResponseCompleted(slimController, t)
	testCheckResponseElse(slimController, t)
	testCheckResponseEmptyId(slimController, t)
}

func testSlimIpErr(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testSlimIpErr", func(t *testing.T) {
		// Test query
		slimController.Post()
	})
}

func testSlimCompressPostErr(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testSlimCompressPostErr", func(t *testing.T) {
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "PathCheck",
			func(_ *controllers.SlimController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

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
		responsePostBody := getResponsePostBody(0, compressInProgress)
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
		responsePostBody := getResponsePostBody(1, "Compress Failed")
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
		responsePostBody := getResponsePostBody(3, "Compress Internal error")
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

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func testAsyCallCompressCompleted(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressCompleted", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		responseGetBody := getResponseGetBody(0, compressCompleted, 1)
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()
		responsePostBody := getResponsePostBody(0, compressInProgress)
		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client,
			url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return &http.Response{Body: responsePostBody}, nil
		})
		defer patch3.Reset()

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func getResponseGetBody(status int, msg string, rate float32) io.ReadCloser {
	var responseGetBodyMap map[string]interface{}
	responseGetBodyMap = make(map[string]interface{})
	responseGetBodyMap["status"] = status
	responseGetBodyMap["msg"] = msg
	responseGetBodyMap["rate"] = rate
	responseGetJson, _ := json.Marshal(responseGetBodyMap)
	responseGetBody := ioutil.NopCloser(bytes.NewReader(responseGetJson))
	return responseGetBody
}

func testAsyCallCompressInProgress(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressInProgress", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		responseGetBody := getResponseGetBody(1, compressInProgress, 0.5)
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

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func testAsyCallCompressFailed(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressFailed", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		responseGetBody := getResponseGetBody(2, "compress failed", 0)
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCheckPostRecordAfterCompress",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ int, _ string,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func testAsyCallCompressNoEnoughSpace(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressNoEnoughSpace", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		responseGetBody := getResponseGetBody(3, "compress NoEnoughSpace", 0)
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCheckPostRecordAfterCompress",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ int, _ string,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func testAsyCallCompressTimeout(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressTimeout", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		responseGetBody := getResponseGetBody(4, "compress Timeout", 0)
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCheckPostRecordAfterCompress",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ int, _ string,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func testAsyCallCompressElse(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressElse", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		responseGetBody := getResponseGetBody(10, "error status", 0)
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCheckPostRecordAfterCompress",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ int, _ string,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func testAsyCallCompressInsertError(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallCompressInsertError", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		responseGetBody := getResponseGetBody(0, "compress success", 1)
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

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func testAsyCallImageOpsGetCheckErr(slimController *controllers.SlimController, t *testing.T, imageFileDb models.ImageDB) {
	t.Run("testAsyCallImageOpsGetCheckErr", func(t *testing.T) {
		// Test query
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		responseGetBody := getResponseGetBody(0, compressCompleted, 1)
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetBody}, nil
		})
		defer patch1.Reset()

		slimController.AsyCallImageOps(client, requestId, LocalIp, imageFileDb, imageId)
	})
}

func testCheckResponseInProgress(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testCheckResponseInProgress", func(t *testing.T) {
		// Test query
		responseGetMapBody := getResponseGetMapBody(4, "check in progress")

		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetMapBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCheckRecordAfterCompress",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ int, _ controllers.CheckStatusResponse,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()
		var imageBasicInfo controllers.ImageBasicInfo
		imageBasicInfo.ImageId = imageId
		imageBasicInfo.StorageMedium = storageMedium
		imageBasicInfo.SaveFileName = saveFileName
		imageBasicInfo.UserId = UserId
		imageBasicInfo.FileName = saveFileName

		var compressInfo controllers.CompressStatusResponse
		compressInfo.Status = 0
		compressInfo.Msg = compressCompleted
		compressInfo.Rate = 1

		slimController.CheckResponse(requestId, imageBasicInfo, compressInfo)
	})
}

func getResponseGetMapBody(status int, msg string) io.ReadCloser {
	var checkInfo controllers.CheckInfo
	checkInfo.Checksum = "111"
	checkInfo.CheckResult = 0
	var responseGetMap map[string]interface{}
	responseGetMap = make(map[string]interface{})
	responseGetMap["status"] = status
	responseGetMap["msg"] = msg
	responseGetMap["checkInfo"] = checkInfo
	responseGetMapJson, _ := json.Marshal(responseGetMap)
	responseGetMapBody := ioutil.NopCloser(bytes.NewReader(responseGetMapJson))
	return responseGetMapBody
}

func testCheckResponseCompleted(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testCheckResponseEmptyId", func(t *testing.T) {
		// Test query
		responseGetMapBody := getResponseGetMapBody(0, "check completed")
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetMapBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCheckRecordAfterCompress",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ int, _ controllers.CheckStatusResponse,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()
		var imageBasicInfo controllers.ImageBasicInfo
		imageBasicInfo.ImageId = imageId
		imageBasicInfo.StorageMedium = storageMedium
		imageBasicInfo.SaveFileName = saveFileName
		imageBasicInfo.UserId = UserId
		imageBasicInfo.FileName = saveFileName

		var compressInfo controllers.CompressStatusResponse
		compressInfo.Status = 0
		compressInfo.Msg = compressCompleted
		compressInfo.Rate = 1

		slimController.CheckResponse(requestId, imageBasicInfo, compressInfo)
	})
}

func testCheckResponseElse(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testCheckResponseEmptyId", func(t *testing.T) {
		// Test query
		responseGetMapBody := getResponseGetMapBody(3, "check else status")
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(client *http.Client, url string) (resp *http.Response, err error) {
			return &http.Response{Body: responseGetMapBody}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(slimController), "InsertOrUpdateCheckRecordAfterCompress",
			func(_ *controllers.SlimController, _ controllers.ImageBasicInfo, _ int, _ controllers.CheckStatusResponse,
				_ controllers.CompressStatusResponse) error {
				return errors.New("error")
			})
		defer patch2.Reset()
		var imageBasicInfo controllers.ImageBasicInfo
		imageBasicInfo.ImageId = imageId
		imageBasicInfo.StorageMedium = storageMedium
		imageBasicInfo.SaveFileName = saveFileName
		imageBasicInfo.UserId = UserId
		imageBasicInfo.FileName = saveFileName

		var compressInfo controllers.CompressStatusResponse
		compressInfo.Status = 0
		compressInfo.Msg = compressCompleted
		compressInfo.Rate = 1

		slimController.CheckResponse(requestId, imageBasicInfo, compressInfo)
	})
}

func testCheckResponseEmptyId(slimController *controllers.SlimController, t *testing.T) {
	t.Run("testCheckResponseEmptyId", func(t *testing.T) {
		// Test query
		var imageBasicInfo controllers.ImageBasicInfo
		imageBasicInfo.ImageId = imageId
		imageBasicInfo.StorageMedium = storageMedium
		imageBasicInfo.SaveFileName = saveFileName
		imageBasicInfo.UserId = UserId
		imageBasicInfo.FileName = saveFileName

		var compressInfo controllers.CompressStatusResponse
		compressInfo.Status = 0
		compressInfo.Msg = compressCompleted
		compressInfo.Rate = 1

		slimController.CheckResponse("", imageBasicInfo, compressInfo)
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

func getResponsePostBody(status int, msg string) io.ReadCloser {
	var responsePostBodyMap map[string]interface{}
	responsePostBodyMap = make(map[string]interface{})
	responsePostBodyMap["status"] = status
	responsePostBodyMap["msg"] = msg
	responsePostBodyMap["requestId"] = requestId
	responsePostJson, _ := json.Marshal(responsePostBodyMap)
	responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))
	return responsePostBody
}
