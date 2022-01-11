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
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestMergeChunkController(t *testing.T) {
	path, extraParams, testDb := prepareTest()
	var c *beego.Controller
	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(c), "ServeJSON", func(*beego.Controller, ...bool) {
		go func() {
			// do nothing
		}()
	})
	defer patch1.Reset()
	mergeChunkController := getMergeChunkController(extraParams, path, testDb)
	testGetMerge(mergeChunkController, t)
	testMergePostPathErr(mergeChunkController, t)
	testMergePostExtensionErr(mergeChunkController, t)
	testMergePost(mergeChunkController, t)
	testMergePostOpenErr(mergeChunkController, t)
	testMergePostIoReadErr(mergeChunkController, t)
	testMergePostNoErr(mergeChunkController, t)
}

func prepareTest() (string, map[string]string, *MockDb) {
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
	return path, extraParams, testDb
}

func testGetMerge(mergeChunkController *controllers.MergeChunkController, t *testing.T) {
	t.Run("testGetMerge", func(t *testing.T) {
		mergeChunkController.Get()
	})
}

func testMergePostPathErr(mergeChunkController *controllers.MergeChunkController, t *testing.T) {
	t.Run("testMergePostPathErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return errors.New("error")
		})
		defer patch1.Reset()
		mergeChunkController.Post()
	})
}

func testMergePostExtensionErr(mergeChunkController *controllers.MergeChunkController, t *testing.T) {
	t.Run("testMergePostExtensionErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		mergeChunkController.Post()
	})
}

func testMergePost(mergeChunkController *controllers.MergeChunkController, t *testing.T) {
	t.Run("testMergePostExtensionErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyFunc(util.ValidateFileExtension, func(_ string) error {
			return nil
		})
		defer patch2.Reset()
		mergeChunkController.Post()
	})
}

func testMergePostOpenErr(mergeChunkController *controllers.MergeChunkController, t *testing.T) {
	t.Run("testMergePostOpenErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyFunc(util.ValidateFileExtension, func(_ string) error {
			return nil
		})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyFunc(os.OpenFile, func(_ string, _ int, _ os.FileMode) (*os.File, error) {
			return nil, errors.New("error")
		})
		defer patch3.Reset()
		mergeChunkController.Post()
	})
}

func testMergePostIoReadErr(mergeChunkController *controllers.MergeChunkController, t *testing.T) {
	t.Run("testMergePostOpenErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyFunc(util.ValidateFileExtension, func(_ string) error {
			return nil
		})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyFunc(ioutil.ReadDir, func(_ string) ([]os.FileInfo, error) {
			return nil, errors.New("error")
		})
		defer patch3.Reset()
		mergeChunkController.Post()
	})
}

func testMergePostNoErr(mergeChunkController *controllers.MergeChunkController, t *testing.T) {
	t.Run("testMergePostNoErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyFunc(util.ValidateFileExtension, func(_ string) error {
			return nil
		})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyFunc(ioutil.ReadDir, func(_ string) ([]os.FileInfo, error) {
			return nil, nil
		})
		defer patch3.Reset()

		patch4 := gomonkey.ApplyFunc(os.OpenFile, func(_ string, _ int, _ os.FileMode) (*os.File, error) {
			return nil, nil
		})
		defer patch4.Reset()

		var responsePostBodyMap map[string]interface{}
		responsePostBodyMap = make(map[string]interface{})
		responsePostBodyMap["status"] = 0
		responsePostBodyMap["msg"] = "Check In Progress"
		responsePostBodyMap["requestId"] = requestId
		responsePostJson, _ := json.Marshal(responsePostBodyMap)
		responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

		patch5 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client, url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return &http.Response{Body: responsePostBody}, nil
		})
		defer patch5.Reset()

		patch6 := gomonkey.ApplyFunc(os.RemoveAll, func(_ string) error {
			return nil
		})
		defer patch6.Reset()

		mergeChunkController.Post()
	})
}

func getMergeChunkController(extraParams map[string]string, path string, testDb dbAdpater.Database) *controllers.MergeChunkController {
	//GET Request
	queryRequest, _ := getHttpRequest(UploadUrl+ZipUri,
		extraParams, "part", path, "GET", []byte(""))

	// Prepare Input
	queryInput := &context.BeegoInput{Context: &context.Context{Request: queryRequest}}
	setParam(queryInput, true)

	// Prepare beego controller
	queryBeegoController := beego.Controller{Ctx: &context.Context{Input: queryInput, Request: queryRequest,
		ResponseWriter: &context.Response{ResponseWriter: httptest.NewRecorder()}},
		Data: make(map[interface{}]interface{})}

	// Create Upload controller with mocked DB and prepared Beego controller
	queryController := &controllers.MergeChunkController{controllers.BaseController{Db: testDb,
		Controller: queryBeegoController}}
	return queryController
}
