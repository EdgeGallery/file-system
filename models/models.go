package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type ImageDB struct {
	ImageId       string `orm:"pk"`
	FileName      string
	UserId        string
	SaveFileName  string
	StorageMedium string
	Url           string
	UploadTime    time.Time `orm:"auto_now_add;type(datetime)"`
}

func init() {
	orm.RegisterModel(new(ImageDB))
}
