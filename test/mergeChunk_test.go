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
	"fileSystem/controllers"
	"fileSystem/models"
	"fileSystem/util"
	"github.com/agiledragon/gomonkey"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMergeGet(t *testing.T) {
	c := getMergeChunkController()
	c.Get()
	// Check for success case wherein the status value will be default i.e. 0
	assert.Equal(t, 0, c.Ctx.ResponseWriter.Status, "Merge get request result received.")
	_ = c.Ctx.ResponseWriter.ResponseWriter.(*httptest.ResponseRecorder)
}

func TestMergePost(t *testing.T) {

	c := getMergeChunkController()

	patch1 := gomonkey.ApplyFunc(util.ValidateSrcAddress, func(_ string) error {
		return nil
	})
	defer patch1.Reset()

	path, _ := os.Getwd()
	path += "/"
	//mockChunkPath := path + "mockChunk" + "/"
	tmpMockChunkPath := path + "mockChunk1" + "/"

	_ = controllers.CreateDirectory(tmpMockChunkPath)

	_, _ = controllers.CopyFile(path+"mockChunk"+"/"+"1.part", tmpMockChunkPath+"1.part")
	_, _ = controllers.CopyFile(path+"mockChunk"+"/"+"2.part", tmpMockChunkPath+"2.part")
	file1, _ := os.Stat(path + "mockChunk1" + "/" + "1.part")
	file2, _ := os.Stat(path + "mockChunk1" + "/" + "2.part")

	patch2 := gomonkey.ApplyMethod(reflect.TypeOf(c), "GetStorageMedium", func(*controllers.MergeChunkController, string) string {
		return path
	})
	defer patch2.Reset()

	fileSlice := []fs.FileInfo{file1, file2}

	patch5 := gomonkey.ApplyFunc(ioutil.ReadDir, func(string) ([]fs.FileInfo, error) {
		return fileSlice, nil
	})
	defer patch5.Reset()

	var responsePostBodyMap map[string]interface{}
	responsePostBodyMap = make(map[string]interface{})
	responsePostBodyMap["status"] = 0
	responsePostBodyMap["msg"] = "Check In Progress"
	responsePostBodyMap["requestId"] = requestId
	responsePostJson, _ := json.Marshal(responsePostBodyMap)
	responsePostBody := ioutil.NopCloser(bytes.NewReader(responsePostJson))

	patch3 := gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Post", func(client *http.Client, url, contentType string, body io.Reader) (resp *http.Response, err error) {
		return &http.Response{Body: responsePostBody}, nil
	})
	defer patch3.Reset()

	patch4 := gomonkey.ApplyMethod(reflect.TypeOf(c), "CronGetCheck", func(*controllers.MergeChunkController, string, string, string, string, string, string) {
		return
	})
	defer patch4.Reset()

	c.Post()

	// Check for success case wherein the status value will be default i.e. 0
	assert.Equal(t, 0, c.Ctx.ResponseWriter.Status, "Merge post request result received.")
	_ = c.Ctx.ResponseWriter.ResponseWriter.(*httptest.ResponseRecorder)

	mactchPath,_ := filepath.Glob(path+"*mock.qcow2")
	for _,v := range mactchPath{
		_ = os.Remove(v)
	}
}

func getMergeChunkController() *controllers.MergeChunkController {
	getBeegoController := beego.Controller{Ctx: &context.Context{ResponseWriter: &context.Response{ResponseWriter: httptest.NewRecorder()}},
		Data: make(map[interface{}]interface{})}
	testDb := &MockDb{
		imageRecords: make(map[string]models.ImageDB),
	}
	c := &controllers.MergeChunkController{BaseController: controllers.BaseController{Db: testDb,
		Controller: getBeegoController}}
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
	c.Ctx.Input.SetParam("identifier", "mockChunk1")
	c.Ctx.Input.SetParam("filename", "mock.qcow2")
	return c
}
