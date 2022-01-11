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
	"errors"
	"fileSystem/controllers"
	"fileSystem/pkg/dbAdpater"
	"fileSystem/util"
	"github.com/agiledragon/gomonkey"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestUploadChunkController(t *testing.T) {
	var c *beego.Controller
	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(c), "ServeJSON", func(*beego.Controller, ...bool) {
		go func() {
			// do nothing
		}()
	})
	defer patch1.Reset()
	path, extraParams, testDb := prepareTest()
	uploadChunkController := getUploadChunkController(extraParams, path, testDb)
	testGet(uploadChunkController, t)
	testPostNoFile(uploadChunkController, t)
	testPost(uploadChunkController, t, path)
	testPostPathErr(uploadChunkController, t)
	testPostGetFileErr(uploadChunkController, t)
	testDelete(uploadChunkController, t)
	testDeletePathErr(uploadChunkController, t)
	testDeleteRemoveErr(uploadChunkController, t)
	testGetStorageMediumA(uploadChunkController, t)
	testGetStorageMediumB(uploadChunkController, t)
	testSaveByIdentifierErr(uploadChunkController, t)
	testSaveByIdentifierCreateErr(uploadChunkController, t)
	testSaveByIdentifier(uploadChunkController, t)
}

func testGet(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testCheckResponseEmptyId", func(t *testing.T) {
		uploadChunkController.Get()
	})
}

func testPostNoFile(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testPostNoFile", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		uploadChunkController.Post()
	})
}

func testPostPathErr(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testPostPathErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return errors.New("error")
		})
		defer patch1.Reset()
		uploadChunkController.Post()
	})
}

func testPostGetFileErr(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testPostGetFileErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()

		var c *beego.Controller
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(c), "GetFile",
			func(*beego.Controller, string) (multipart.File, *multipart.FileHeader, error) {
				return nil, nil, errors.New("error")
			})
		defer patch2.Reset()
		uploadChunkController.Post()
	})
}

func testPost(uploadChunkController *controllers.UploadChunkController, t *testing.T, path string) {
	t.Run("testPost", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		uploadChunkController.Ctx.Input.SetParam("file", path+"/mockChunk/1.part")
		uploadChunkController.Post()
	})
}

func testDelete(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testPost", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		uploadChunkController.Delete()
	})
}

func testDeletePathErr(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testDeletePathErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return errors.New("error")
		})
		defer patch1.Reset()
		uploadChunkController.Delete()
	})
}

func testDeleteRemoveErr(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testDeleteRemoveErr", func(t *testing.T) {
		patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
			return nil
		})
		defer patch1.Reset()
		patch6 := gomonkey.ApplyFunc(os.RemoveAll, func(_ string) error {
			return errors.New("error")
		})
		defer patch6.Reset()
		uploadChunkController.Delete()
	})
}

func testGetStorageMediumA(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testGetStorageMedium", func(t *testing.T) {
		uploadChunkController.GetStorageMedium("A")
	})
}

func testGetStorageMediumB(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testGetStorageMedium", func(t *testing.T) {
		uploadChunkController.GetStorageMedium("B")
	})
}

func testSaveByIdentifierErr(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testSaveByIdentifierErr", func(t *testing.T) {
		err := uploadChunkController.SaveByIdentifier("A", "saveFileName", "identifier")
		assert.Error(t, err, "sorry, this storage medium is not supported right now")
	})
}

func testSaveByIdentifierCreateErr(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testSaveByIdentifierCreateErr", func(t *testing.T) {
		patch4 := gomonkey.ApplyFunc(controllers.CreateDirectory, func(_ string) error {
			return errors.New("create file error")
		})
		defer patch4.Reset()
		err := uploadChunkController.SaveByIdentifier("0", "saveFileName", "identifier")
		assert.Error(t, err, "create file error")
	})
}

func testSaveByIdentifier(uploadChunkController *controllers.UploadChunkController, t *testing.T) {
	t.Run("testSaveByIdentifier", func(t *testing.T) {
		var c *beego.Controller
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(c), "SaveToFile", func(*beego.Controller, string, string) error {
			return errors.New("error")
		})
		defer patch1.Reset()
		err := uploadChunkController.SaveByIdentifier("0", "saveFileName", "identifier")
		assert.Error(t, err, "error")
	})
}

func getUploadChunkController(extraParams map[string]string, path string, testDb dbAdpater.Database) *controllers.UploadChunkController {
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
	queryController := &controllers.UploadChunkController{controllers.BaseController{Db: testDb,
		Controller: queryBeegoController}}
	return queryController
}
