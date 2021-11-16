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

// @Title   util
// @Description  util pkg
// @Author  GuoZhen Gao (2021/6/30 10:40)
package util

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/go-playground/validator/v10"
	"os"
	"path/filepath"
)

const (
	BadRequest                int    = 400
	StatusInternalServerError int    = 500
	StatusNotFound            int    = 404
	ClientIpaddressInvalid           = "clientIp address is invalid"
	LastInsertIdNotSupported  string = "LastInsertId is not supported by this driver"
	FileRecord                string = "Add file record: %+v"
	FailedToDecompress               = "Failed to decompress zip file"
	FailedToDeleteCache              = "Failed to delete cache file"
	FailedToUnmarshal                = "failed to unmarshal request"
	FailedToInsertDataToDB    string = "fail to insert imageID, filename, userID to database"
	FailToInsertRequestCheck  string = "fail to insert request imageOps check to db"
	OriginalNameIs                   = "originalName is"
	FailToRecordToDB                 = "Failed to save file record to database."
	TypeNotSupport                   = "This image cannot be slimmed because the type of image is not supported."
	ImageSlimming                    = "The image file is being slimmed. No need to slim again."
	ImageSlimmed                     = "The image file has already been slimmed. No need to slim again. Pls request to check directly"
	SlimExitNoSpace                  = "Compress exiting because of No enough space left"
	Default                   string = "default"
	MaxFileNameSize                  = 128
	MaxIPVal                         = 255
	MaxAppPackageFile         int64  = 536870912000 //fix file size here
	Operation                        = "] Operation ["
	Resource                         = " Resource ["
	LocalStoragePath          string = "/usr/app/vmImage/"
	FormFile                  string = "file"
	UserId                    string = "userId"
	Priority                  string = "priority"
	Part                      string = "part"
	Identifier                string = "identifier"
	FileName                  string = "filename"
	DriverName                string = "postgres"
	SslMode                   string = "disable"

	DbImageId           string = "image_id"
	DbFileName          string = "file_name"
	DbUserId            string = "user_id"
	DbSaveFileName      string = "save_file_name"
	DbStorageMedium     string = "storage_medium"
	DbUploadTime        string = "upload_time"
	DbSlimStatus        string = "slim_status"
	DbRequestIdCheck    string = "request_id_check"
	DbRequestIdCompress string = "request_id_compress"
	DbChecksum          string = "checksum"
	DbCheckResult       string = "check_result"
	DbCheckMsg          string = "check_msg"
	DbCheckStatus       string = "check_status"
	DbImageEndOffset    string = "image_end_offset"
	DbCheckErrors       string = "check_errors"
	DbFormat            string = "format"
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

// Validate file extension
func ValidateFileExtension(fileName string) error {
	extension := filepath.Ext(fileName)
	if extension != ".zip" && extension != ".qcow2" && extension != ".img" && extension != ".iso" {
		return errors.New("file extension is not supported")
	}
	return nil
}

// Get app configuration
func GetAppConfig(k string) string {
	return beego.AppConfig.String(k)
}

// Get db user
func GetDbUser() string {
	dbUser := os.Getenv("POSTGRES_USERNAME")
	return dbUser
}

// Get database name
func GetDbName() string {
	dbName := os.Getenv("POSTGRES_DB_NAME")
	return dbName
}

// Get database host
func GetDbHost() string {
	dbHost := os.Getenv("POSTGRES_HOST")
	return dbHost
}

// Get database port
func GetDbPort() string {
	dbPort := os.Getenv("POSTGRES_PORT")
	return dbPort
}

// Clear byte array from memory
func ClearByteArray(data []byte) {
	for i := 0; i < len(data); i++ {
		data[i] = 0
	}
}
