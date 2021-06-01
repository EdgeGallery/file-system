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
	"errors"
	"github.com/astaxie/beego"
	"github.com/go-playground/validator/v10"
	"path/filepath"
)

const (
	BadRequest                int = 400
	StatusUnauthorized        int = 401
	StatusInternalServerError int = 500
	StatusNotFound            int = 404
	StatusForbidden           int = 403

	ClientIpaddressInvalid        = "clientIp address is invalid"
	Default                string = "default"
	MaxFileNameSize               = 128
	MaxAppPackageFile      int64  = 536870912000 //fix file size here
	Operation                     = "] Operation ["
	Resource                      = " Resource ["

	LocalStoragePath       string = "/usr/vmImage/"
	FormFile               string = "file"
	UserId                 string = "userId"
	Priority               string = "priority"
)

// Validate file size
func ValidateFileSize(fileSize int64, maxFileSize int64) error {
	if fileSize < maxFileSize {
		return nil
	}
	return errors.New("invalid file, file size is larger than max size")
}

// Validate source address
func ValidateSrcAddress(id string) error {
	if id == "" {
		return errors.New("require ip address")
	}

	validate := validator.New()
	err := validate.Var(id, "required,ipv4")
	if err != nil {
		return validate.Var(id, "required,ipv6")
	}
	return nil
}

// Validate file extenstion
func ValidateFileExtensionZip(fileName string) error {
	extension := filepath.Ext(fileName)
	if extension != ".zip" {
		return errors.New("file extension is not zip")
	}
	return nil
}

// Get app configuration
func GetAppConfig(k string) string {
	return beego.AppConfig.String(k)
}
