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
	"os"
	"path/filepath"
	"regexp"
)

const (
	BadRequest                int = 400
	StatusUnauthorized        int = 401
	StatusInternalServerError int = 500
	StatusNotFound            int = 404
	StatusForbidden           int = 403

	ClientIpaddressInvalid          = "clientIp address is invalid"
	LastInsertIdNotSupported string = "LastInsertId is not supported by this driver"
	FailedToDecompress              = "Failed to decompress zip file"
	FailedToDeleteCache             = "Failed to delete cache file"
	Default                  string = "default"
	MaxFileNameSize                 = 128
	MaxAppPackageFile        int64  = 536870912000 //fix file size here
	Operation                       = "] Operation ["
	Resource                        = " Resource ["
	SingleFile                      = 1
	TooManyFile                     = 1024
	FailedToMakeDir                 = "failed to make directory"
	TooBig                          = 0x6400000
	SingleFileTooBig                = 0x6400000
	LocalStoragePath         string = "/usr/app/vmImage/"
	FormFile                 string = "file"
	UserId                   string = "userId"
	Priority                 string = "priority"
	DriverName               string = "postgres"
	SslMode                  string = "disable"
	minPasswordSize         = 8
	maxPasswordSize         = 16
	maxPasswordCount        = 2
	singleDigitRegex string = `\d`
	lowerCaseRegex   string = `[a-z]`
	upperCaseRegex   string = `[A-Z]`
	specialCharRegex string = `['~!@#$%^&()-_=+\|[{}\];:'",<.>/?]`
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
	if extension != ".zip" && extension != ".qcow2" && extension != ".img" &&extension!=".iso" {
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

// Validate db parameters
func ValidateDbParams(dbPwd string) (bool, error) {
	dbPwdBytes := []byte(dbPwd)
	dbPwdIsValid, validateDbPwdErr := ValidatePassword(&dbPwdBytes)
	if validateDbPwdErr != nil || !dbPwdIsValid {
		return dbPwdIsValid, validateDbPwdErr
	}
	return true, nil
}

// Validate password
func ValidatePassword(password *[]byte) (bool, error) {
	if len(*password) >= minPasswordSize && len(*password) <= maxPasswordSize {
		// password must satisfy any two conditions
		pwdValidCount := GetPasswordValidCount(password)
		if pwdValidCount < maxPasswordCount {
			return false, errors.New("password must contain at least two types of the either one lowercase" +
				" character, one uppercase character, one digit or one special character")
		}
	} else {
		return false, errors.New("password must have minimum length of 8 and maximum of 16")
	}
	return true, nil
}

// To get password valid count
func GetPasswordValidCount(password *[]byte) int {
	var pwdValidCount = 0
	pwdIsValid, err := regexp.Match(singleDigitRegex, *password)
	if pwdIsValid && err == nil {
		pwdValidCount++
	}
	pwdIsValid, err = regexp.Match(lowerCaseRegex, *password)
	if pwdIsValid && err == nil {
		pwdValidCount++
	}
	pwdIsValid, err = regexp.Match(upperCaseRegex, *password)
	if pwdIsValid && err == nil {
		pwdValidCount++
	}
	// space validation for password complexity is not added
	// as jwt decrypt fails if space is included in password
	pwdIsValid, err = regexp.Match(specialCharRegex, *password)
	if pwdIsValid && err == nil {
		pwdValidCount++
	}
	return pwdValidCount
}
