package main

import (
	_ "fileSystem/models"
	_ "fileSystem/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"net/http"
)






func main() {

	beego.InsertFilter("*", beego.BeforeRouter,cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "X-Requested-With", "Content-Type", "Accept"},
		ExposeHeaders: []string{"Content-Length"},
		AllowCredentials: true,
	}))

	beego.ErrorHandler("429", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	})


	beego.Run()
}

