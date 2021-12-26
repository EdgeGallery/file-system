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
	"fileSystem/models"
	"github.com/agiledragon/gomonkey"
	"github.com/astaxie/beego"
	"os"
	"reflect"
	"testing"
)

func TestControllerZipSuccess(t *testing.T) {

	// Common steps
	// Setting file path
	// return filesystem/test的目录地址
	path, _ := os.Getwd()
	path += "/mockImage.zip"

	// Setting extra parameters
	extraParams := map[string]string{
		UserIdKey:   UserId,
		PriorityKey: Priority,
	}

	testDb := &MockDb{
		imageRecords: models.ImageDB{},
	}

	var c *beego.Controller
	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(c), "ServeJSON", func(*beego.Controller, ...bool) {
		go func() {
			// do nothing
		}()
	})
	defer patch1.Reset()

	testUploadPostValidateSrcAddressErr(t, extraParams, path, testDb)
	testUploadPostValidateSrcAddress(t, extraParams, path, testDb)

}