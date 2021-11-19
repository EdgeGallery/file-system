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
	"fileSystem/models"
	"fileSystem/pkg/dbAdpater"
	"fileSystem/util"
	"github.com/agiledragon/gomonkey"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

var (
	BaseUrl     string = "http://edgegallery:9500/image-management/v1"
	ImageId     string = "94d6e70d-51f7-4b0d-965f-59dca2c3002c"
	UserIdKey   string = "userId"
	UserId      string = "71ea862b-5806-4196-bce3-434bf9c95b18"
	PriorityKey string = "priority"
	Priority    string = "0"
	err                = errors.New("error")
	UploadBody  string = "{\n\"userId\":" +
		"\"e921ce54-82c8-4532-b5c6-8516cf75f7a771ea862b-5806-4196-bce3-434bf9c95b18\",\n\"tenantId\":" +
		"\"e921ce54-82c8-4532-b5c6-8516cf75f7a7\",\n\"appInstanceId\":\"71ea862b-5806-4196-bce3-434bf9c95b18\",\n\"" +
		"appName\":\"abcd\",\n\"appSupportMp1\": true,\n\"appTrafficRule\": [\n{\n\"trafficRuleId\":\"TrafficRule1\",\n\"" +
		"filterType\":\"FLOW\",\n\"priority\": 1,\n\"action\":\"DROP\",\n\"trafficFilter\":[\n{\n\"trafficFilterId\":" +
		"\"75256a74-adb9-4c6d-8246-9773dfd5f6df\",\n\"srcAddress\":[\n\"192.168.1.1/28\",\n\"192.168.1.2/28\"\n],\n\"" +
		"srcPort\":[\n\"6666666666\"\n],\n\"dstAddress\":[\n\"192.168.1.1/28\"\n],\n\"dstPort\":[\n\"6666666666\"\n],\n\"" +
		"protocol\":[\n\"TCP\"\n],\n\"qCI\": 1,\n\"dSCP\":0,\n\"tC\": 1,\n\"tag\":[\n\"1\"\n],\n\"srcTunnelAddress\":" +
		"[\n\"1.1.1.1/24\"\n],\n\"dstTunnelAddress\":[\n\"1.1.1.1/24\"\n],\n\"srcTunnelPort\":[\n\"65536\"\n],\n\"" +
		"dstTunnelPort\":[\n\"65537\"\n]\n}\n],\n\"dstInterface\":[\n{\n\"dstInterfaceId\":" +
		"\"caf2dab7-0c20-4fe7-ac72-a7e204a309d2\",\n\"interfaceType\":\"\",\n\"srcMacAddress\":\"\",\n\"dstMacAddress\":" +
		"\"\",\n\"dstIpAddress\":\"\",\n\"TunnelInfo\":{\n\"tunnelInfoId\":\"461ceb53-291c-422c-9cbe-27f40e4ad2b3\",\n\"" +
		"tunnelType\":\"\",\n\"tunnelDstAddress\":\"\",\n\"tunnelSrcAddress\":\"\",\n\"tunnelSpecificData\":" +
		"\"\"\n}\n}\n]\n}\n],\n\"appDnsRule\":[\n{\n\"dnsRuleId\":\"dnsRule4\",\n\"domainName\":\"www.example.com\",\n\"" +
		"ipAddressType\":\"IP_V4\",\n\"ipAddress\":\"192.0.2.0\",\n\"ttl\":30\n}\n],\n\"Origin\":\"MEPM\"\n}"

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

	testUploadGet(t, extraParams, "", testDb)

	testUploadPostValidateSrcAddressErr(t, extraParams, path, testDb)
	testUploadPostValidateSrcAddress(t, extraParams, path, testDb)
}

func testUploadGet(t *testing.T, extraParams map[string]string, path string, testDb dbAdpater.Database) {

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
}

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

func setParam(ctx *context.BeegoInput, isZip bool) {
	if isZip {
		ctx.SetParam("isZip", "true")
	}
	ctx.SetParam(":imageId", ImageId)
}
