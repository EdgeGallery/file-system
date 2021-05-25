package routers

import (
	"fileSystem/controllers"
	"fileSystem/pkg/dbAdpater"
	"github.com/astaxie/beego"

	"os"
)

func init() {
	adapter := initDbAdapter()

/*	ns := beego.NewNamespace("/imagemangement/v1/",
		beego.NSInclude(
			&controllers.UploadController{controllers.BaseController{Db: adapter}},
			//&controllers.FileDownloadController{controllers.BaseController{Db: adapter}},
		),
	)
	beego.AddNamespace(ns)*/

	beego.Router("/imagemangement/v1/upload",&controllers.UploadController{controllers.BaseController{Db: adapter}})
	beego.Router("/imagemangement/v1/download", &controllers.DownloadController{controllers.BaseController{Db: adapter}})

}

// Init Db adapter
func initDbAdapter() (pgDb dbAdpater.Database) {
	adapter, err := dbAdpater.GetDbAdapter()
	if err != nil {
		os.Exit(1)
	}
	return adapter
}