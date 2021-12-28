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
	path += "/mockImage.qcow2"

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

	testDownloadNoErr(t, extraParams, path, testDb)
	testDownloadIPError(t, extraParams, path, testDb)
	testDownloadPathError(t, extraParams, path, testDb)
	testDownloadPathOk(t, extraParams, path, testDb)
	testDownloadCopyErr(t, extraParams, path, testDb)
	testDownloadOsOpenErr(t, extraParams, path, testDb)
	testDownloadCompressErr(t, extraParams, path, testDb)

}

func testDownloadIPError(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {

	t.Run("testDownloadIPError", func(t *testing.T) {
		//GET Request
		queryRequest, _ := getHttpRequest(UploadUrl+
			"/94d6e70d-51f7-4b0d-965f-59dca2c3002c/action/download",
			extraParams, "file", path, "GET", []byte(""))

		// Prepare Input
		queryInput := &context.BeegoInput{Context: &context.Context{Request: queryRequest}}
		setParam(queryInput, false)

		// Prepare beego controller
		queryBeegoController := beego.Controller{Ctx: &context.Context{Input: queryInput, Request: queryRequest,
			ResponseWriter: &context.Response{ResponseWriter: httptest.NewRecorder()}},
			Data: make(map[interface{}]interface{})}

		// Create Upload controller with mocked DB and prepared Beego controller
		queryController := &controllers.DownloadController{controllers.BaseController{Db: testDb,
			Controller: queryBeegoController}}

		queryController.Get()
	})
}

func testDownloadPathError(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {
	t.Run("testDownloadPathError", func(t *testing.T) {
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

		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		// Test query
		queryController.Get()
	})
}

func testDownloadPathOk(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {
	t.Run("testDownloadPathOk", func(t *testing.T) {
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

func testDownloadCopyErr(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {
	t.Run("testDownloadPathOk", func(t *testing.T) {
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

func testDownloadOsOpenErr(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {
	t.Run("testDownloadPathOk", func(t *testing.T) {
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

func testDownloadCompressErr(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {
	t.Run("testDownloadCompressErr", func(t *testing.T) {
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

func testDownloadNoErr(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {
	t.Run("testDownloadNoErr", func(t *testing.T) {
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
