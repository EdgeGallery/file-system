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

	ClientIpaddressInvalid          = "clientIp address is invalid"
	Default                  string = "default"
	MaxFileNameSize                 = 128
	MaxAppPackageFile        int64  = 536870912000   //fix file size here
	Operation            = "] Operation ["
	Resource             = " Resource ["


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

