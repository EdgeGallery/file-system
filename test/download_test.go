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
	"fileSystem/controllers"
	"fileSystem/models"
	"fileSystem/pkg/dbAdpater"
	"fileSystem/util"
	"github.com/agiledragon/gomonkey"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestDownloadController(t *testing.T) {

	// Common steps
	// Setting file path
	// return filesystem/test的目录地址
	path, _ := os.Getwd()

	// Setting extra parameters
	extraParams := map[string]string{
		UserIdKey:   UserId,
		PriorityKey: Priority,
	}

	fileRecord := models.ImageDB{
		ImageId:       imageId,
		FileName:      util.FileName,
		UserId:        UserId,
		SaveFileName:  saveFileName,
		StorageMedium: storageMedium,
		SlimStatus:    2,
	}

	testDb := &MockDb{
		imageRecords: fileRecord,
	}

	var c *beego.Controller
	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(c), "ServeJSON", func(*beego.Controller, ...bool) {
		go func() {
			// do nothing
		}()
	})
	defer patch1.Reset()
	queryController := getController(extraParams, path, testDb)
	testDownloadIPError(queryController, t)
	testDownloadPathError(queryController, t)
	testDownloadPathOk(queryController, t)
	testDownloadCopyErr(queryController, t)
	testDownloadOsOpenErr(queryController, t)
	testDownloadCompressErr(queryController, t)
	testDownloadNoErr(queryController, t)
	os.Remove(path + "/.zip")
}

func testDownloadIPError(queryController *controllers.DownloadController, t *testing.T) {

	t.Run("testDownloadIPError", func(t *testing.T) {
		queryController.Get()
	})
}

func testDownloadPathError(queryController *controllers.DownloadController, t *testing.T) {
	t.Run("testDownloadPathError", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		// Test query
		queryController.Get()
	})
}

func testDownloadPathOk(queryController *controllers.DownloadController, t *testing.T) {
	t.Run("testDownloadPathOk", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(queryController), "PathCheck",
			func(_ *controllers.DownloadController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(queryController.Db), "QueryTable",
			func(_ *MockDb, _ string, _ interface{}, _ string, _ ...interface{}) (num int64, err error) {
				return 0, nil
			})
		defer patch3.Reset()

		// Test query
		queryController.Get()
	})
}

func testDownloadCopyErr(queryController *controllers.DownloadController, t *testing.T) {
	t.Run("testDownloadPathOk", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(queryController), "PathCheck",
			func(_ *controllers.DownloadController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(queryController.Db), "QueryTable",
			func(_ *MockDb, _ string, _ interface{}, _ string, _ ...interface{}) (num int64, err error) {
				return 0, nil
			})
		defer patch3.Reset()

		patch4 := gomonkey.ApplyFunc(controllers.CreateDirectory, func(_ string) error {
			return nil
		})
		defer patch4.Reset()

		// Test query
		queryController.Get()

	})
}

func testDownloadOsOpenErr(queryController *controllers.DownloadController, t *testing.T) {
	t.Run("testDownloadPathOk", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(queryController), "PathCheck",
			func(_ *controllers.DownloadController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(queryController.Db), "QueryTable",
			func(_ *MockDb, _ string, _ interface{}, _ string, _ ...interface{}) (num int64, err error) {
				return 0, nil
			})
		defer patch3.Reset()

		patch4 := gomonkey.ApplyFunc(controllers.CreateDirectory, func(_ string) error {
			return nil
		})
		defer patch4.Reset()

		patch5 := gomonkey.ApplyFunc(controllers.CopyFile, func(_ string, _ string) (int64, error) {
			return 0, nil
		})
		defer patch5.Reset()

		// Test query
		queryController.Get()
	})
}

func testDownloadCompressErr(queryController *controllers.DownloadController, t *testing.T) {
	t.Run("testDownloadCompressErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(queryController), "PathCheck",
			func(_ *controllers.DownloadController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(queryController.Db), "QueryTable",
			func(_ *MockDb, _ string, _ interface{}, _ string, _ ...interface{}) (num int64, err error) {
				return 0, nil
			})
		defer patch3.Reset()

		patch4 := gomonkey.ApplyFunc(controllers.CreateDirectory, func(_ string) error {
			return nil
		})
		defer patch4.Reset()

		patch5 := gomonkey.ApplyFunc(controllers.CopyFile, func(_ string, _ string) (int64, error) {
			return 0, nil
		})
		defer patch5.Reset()

		patch6 := gomonkey.ApplyFunc(os.Open, func(_ string) (*os.File, error) {
			return nil, nil
		})
		defer patch6.Reset()

		// Test query
		queryController.Get()
	})
}

func testDownloadNoErr(queryController *controllers.DownloadController, t *testing.T) {
	t.Run("testDownloadNoErr", func(t *testing.T) {
		patch6 := gomonkey.ApplyFunc(os.Open, func(_ string) (*os.File, error) {
			return nil, nil
		})
		defer patch6.Reset()
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(queryController), "PathCheck",
			func(_ *controllers.DownloadController, _ string) bool {
				return true
			})
		defer patch2.Reset()

		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(queryController.Db), "QueryTable",
			func(_ *MockDb, _ string, _ interface{}, _ string, _ ...interface{}) (num int64, err error) {
				return 0, nil
			})
		defer patch3.Reset()

		patch4 := gomonkey.ApplyFunc(controllers.CreateDirectory, func(_ string) error {
			return nil
		})
		defer patch4.Reset()

		patch5 := gomonkey.ApplyFunc(controllers.CopyFile, func(_ string, _ string) (int64, error) {
			return 0, nil
		})
		defer patch5.Reset()

		patch7 := gomonkey.ApplyFunc(controllers.Compress, func(_ []*os.File, _ string) error {
			return nil
		})
		defer patch7.Reset()

		patch8 := gomonkey.ApplyMethod(reflect.TypeOf(queryController.Ctx.Output), "Download",
			func(_ *context.BeegoOutput, _ string, _ ...string) {
				return
			})
		defer patch8.Reset()

		// Test query
		queryController.Get()
	})
}

func testDownloadNotZip(queryController *controllers.DownloadController, t *testing.T) {

	t.Run("testDownloadIPError", func(t *testing.T) {
		queryController.Get()
	})
}

func getController(extraParams map[string]string, path string, testDb dbAdpater.Database) *controllers.DownloadController {
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
	queryController := &controllers.DownloadController{controllers.BaseController{Db: testDb,
		Controller: queryBeegoController}}
	return queryController
}
