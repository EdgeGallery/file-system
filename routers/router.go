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

// @Title   routers
// @Description  rout api
// @Author  GuoZhen Gao (2021/6/30 10:40)
package routers

import (
	"fileSystem/controllers"
	"fileSystem/pkg/dbAdpater"
	"github.com/astaxie/beego"

	"os"
)

func init() {
	adapter := initDbAdapter()

	beego.Router("/image-management/v1/images",&controllers.UploadController{controllers.BaseController{Db: adapter}})
	beego.Router("/image-management/v1/images/:imageId/action/download", &controllers.DownloadController{controllers.BaseController{Db: adapter}})
	beego.Router("/image-management/v1/images/:imageId",&controllers.ImageController{controllers.BaseController{Db: adapter}})
	beego.Router("/image-management/v1/images/upload",&controllers.UploadChunkController{controllers.BaseController{Db: adapter}})
	beego.Router("/image-management/v1/images/merge",&controllers.MergeChunkController{controllers.BaseController{Db: adapter}})
	beego.Router("/image-management/v1/images/:imageId/action/slim",&controllers.SlimController{controllers.BaseController{Db: adapter}})

}

// Init Db adapter
func initDbAdapter() (pgDb dbAdpater.Database) {
	adapter, err := dbAdpater.GetDbAdapter()
	if err != nil {
		os.Exit(1)
	}
	return adapter
}