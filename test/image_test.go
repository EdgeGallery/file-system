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
	"fileSystem/util"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImageGet(t *testing.T) {
	c := getImageController()
	c.Ctx.Input.SetParam(":imageId", imageId)
	c.Get()
	// Check for success case wherein the status value will be default i.e. 0
	assert.Equal(t, 0, c.Ctx.ResponseWriter.Status, "Image get request result received.")
	_ = c.Ctx.ResponseWriter.ResponseWriter.(*httptest.ResponseRecorder)
}

func TestImageGetImageError(t *testing.T) {
	c := getImageController()
	c.Get()
	// Check for success case wherein the status value will be default i.e. 0
	assert.Equal(t, util.StatusNotFound, c.Ctx.ResponseWriter.Status, "Image get request result received.")
	_ = c.Ctx.ResponseWriter.ResponseWriter.(*httptest.ResponseRecorder)
}

func TestImageDelete(t *testing.T) {
	c := getImageController()
	c.Ctx.Input.SetParam(":imageId", imageId)
	c.Delete()
	// Check for success case wherein the status value will be default i.e. 0
	assert.Equal(t, 0, c.Ctx.ResponseWriter.Status, "Image get request result received.")
	_ = c.Ctx.ResponseWriter.ResponseWriter.(*httptest.ResponseRecorder)
}

func getImageController() *controllers.ImageController {
	getBeegoController := beego.Controller{Ctx: &context.Context{ResponseWriter: &context.Response{ResponseWriter:
	httptest.NewRecorder()}},
		Data: make(map[interface{}]interface{})}

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
	c := &controllers.ImageController{BaseController: controllers.BaseController{Db: testDb,
		Controller: getBeegoController}}
	c.Init(context.NewContext(), "", "", nil)
	req, err := http.NewRequest("GET", "http://127.0.0.1", strings.NewReader(""))
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
	return c
}
