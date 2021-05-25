package controllers

import (
	"fileSystem/models"
	"fileSystem/util"
	log "github.com/sirupsen/logrus"
)

//下载文件
type DownloadController struct {
	BaseController
}


func (this *DownloadController) Get() {
	log.Info("Download get request received.")

	clientIp := this.Ctx.Input.IP()
	err := util.ValidateSrcAddress(clientIp)
	if err != nil {
		this.HandleLoggingForError(clientIp, util.BadRequest, util.ClientIpaddressInvalid)
		return
	}

	this.displayReceivedMsg(clientIp)

	var imageFileDb models.ImageDB

	imageId := this.Ctx.Input.Query("imageId")

	_, err = this.Db.QueryTable("image_d_b", &imageFileDb, "image_id__exact", imageId)

	//err = this.Db.QueryForDownload("image_d_b", &imageFileDb, imageId) //表名
	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusNotFound, "fail to query database")
		return
	}

	filePath := imageFileDb.StorageMedium
	err = createDirectory(filePath)
	if err != nil {
		log.Error("failed to create file path" + filePath)
	}

	fileName := imageFileDb.SaveFileName
	originalName := imageFileDb.FileName

	downloadPath := filePath + fileName

	//第一个参数是文件的地址，第二个参数是下载显示的文件的名称
	//this.Ctx.Output.Download("static/healthcheck工作计划.xlsx", "1.xlsx")

	//加文件下载路径
	this.Ctx.Output.Download(downloadPath,originalName)

	/*this.Ctx.WriteString("download success")
	log.Info("save file to " + downloadPath)*/
	//this.Ctx.Output.Download("/usr/vmImage/1.zip", "download.zip")

	/*downloadResp, err := json.Marshal(map[string]string{
		"imageId":    imageId,
		"uploadTime": time.Now().Format("2006-01-02 15:04:05"),
		"download":   downloadPath})

	if err != nil {
		this.HandleLoggingForError(clientIp, util.StatusInternalServerError, "fail to return download details")
		return
	}
	_, _ = this.Ctx.ResponseWriter.Write(downloadResp)*/
}
