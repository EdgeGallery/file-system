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

// @Title  models
// @Description  control database
// @Author  GuoZhen Gao (2021/6/30 10:40)
package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

// ImageDB   Define the database type
type ImageDB struct {
	ImageId        string `orm:"pk"`
	FileName       string
	UserId         string
	SaveFileName   string
	StorageMedium  string
	UploadTime     time.Time `orm:"auto_now_add;type(datetime)"`
	SlimStatus     int       //[0,1,2,3]  成功/失败/未瘦身/瘦身中
	RequestId      string
	Checksum       string
	ImageEndOffset string
	CheckErrors    string
	format         string
}

func init() {
	orm.RegisterModel(new(ImageDB))
}
