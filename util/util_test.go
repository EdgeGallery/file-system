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

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateFileSizeSuccess(t *testing.T) {
	err := ValidateFileSize(10, 100)
	assert.Nil(t, err, "TestValidateFileSizeSuccess execution result")
}

func TestValidateFileSizeInvalid(t *testing.T) {
	err := ValidateFileSize(100, 10)
	assert.Error(t, err, "TestValidateFileSizeInvalid execution result")
}

func TestValidateSrcAddressNull(t *testing.T) {
	err := ValidateSrcAddress("")
	assert.Error(t, err,"TestValidateSrcAddressNull execution result")
}

func TestValidateSrcAddressIPv4Success(t *testing.T) {
	err := ValidateSrcAddress("127.0.0.1")
	assert.Nil(t, err,"TestValidateSrcAddressIPv4Success execution result")
}

func TestValidateSrcAddressIPv6Success(t *testing.T) {
	err := ValidateSrcAddress("1:1:1:1:1:1:1:1")
	assert.Nil(t, err,"TestValidateSrcAddressIPv6Success execution result")
}

func TestValidateFileExtensionInvalid(t *testing.T) {
	err := ValidateFileExtension("x.txt")
	assert.Error(t, err, "TestValidateFileExtensionInvalid execution result")
}

func TestValidateFileExtensionSuccess(t *testing.T) {
	err := ValidateFileExtension("x.qcow2")
	assert.Nil(t, err, "TestValidateFileExtensionSuccess execution result")
}

func TestGetAppConfig(_ *testing.T) {
	appConfig := "appConfig"
	GetAppConfig(appConfig)
}

func TestGetDbUser(t *testing.T) {
	err := GetDbUser()
	assert.Equal(t, "", err, "TestGetDbUser execution result")
}

func TestGetDbName(t *testing.T) {
	err := GetDbName()
	assert.Equal(t, "", err, "TestGetDbName execution result")
}

func TestGetDbHost(t *testing.T) {
	err := GetDbHost()
	assert.Equal(t, "", err, "TestGetDbHost execution result")
}

func TestGetDbPort(t *testing.T) {
	err := GetDbPort()
	assert.Equal(t, "", err, "TestGetDbPort execution result")
}

func TestClearByteArray(t *testing.T) {
	data := []byte{1,2,3}
    ClearByteArray(data)
}




