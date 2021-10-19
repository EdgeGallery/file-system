# EdgeGallery file-system



## Overview

   As a middleware, file system provides file management services for uploading, downloading, querying and deleting for multiple modules. The image file slimming function is added in v1.3



## Detailed introduction

- upload

  The image file itself can be selected for uploading the file. If the image file is too large, it is recommended to compress it into a zip file for uploading, or it is recommended to upload it in pieces. When uploading zip, the. qcow2 file needs to be compressed in the folder.

- download

  Different formats can be selected when downloading files according to imageId, which can be controlled by QUERY parameters

- query

 Query file details based on imageId

- delete

  Delete stored files and data information according to imageId

- Slim

  Call the imageops function of another container in the same pod to assist in the image slimming operation. The specific flow chart is as follows:

![image-20211019163303625](file://C:\Users\Administrator\Desktop\image-20211019163303625.png?lastModify=1634634782)

## API Defination

|                     | Method | URL                                                   | form-data                                                    | Response                                                     | API Instruction                                              |
| ------------------- | :----: | ----------------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| upload              |  POST  | /image-management/v1/images                           | userId: the ID of user<br/>file: chosen file<br/>priority: Storage priority | {<br/>imageId:"string"<br/>fileName:"string"<br/>uploadTime:"string"<br/>userId:"string"<br/>storageMedium:"string"<br/>msg:"string"<br/>checkStatus:int<br/>slimStatus:int<br/>} | slimStatus:[0,1,2,3]stands for Not slimmed yet / slimming down / success / failure <br/>Optional upload image file format：.zip/.qcow2/.img/.iso, ；<br/>***priority*** generally select 0；When uploading the zip, the upper layer of the image file should be wrapped with a layer of folders |
| download            |  GET   | /image-management/v1/images/{imageId}/action/download | None                                                         | file                                                         | The format of the download image file is optional. When the query is /?isZip=true, the download format is .zip; when the query is not included, the image file itself is downloaded |
| query               |  GET   | /image-management/v1/images/{imageId}                 | None                                                         | {imageId:"string"<br/>file:"string"<br/>uploadTime:"string"<br/>userId:"string"<br/>storageMedium:"string"} | Query file details based on imageId                          |
| delete              | DELETE | /image-management/v1/images/{imageId}                 | None                                                         | 删除成功: delete success/<br/>删除失败: error                | Delete local files based on imageId                          |
| slim                |  POST  | /image-management/v1/images/{imageId}/action/slim     | None                                                         | compress in progress/ <br/>compress failed                   | Compress the image file according to ImageID                 |
| upload chunk        |  POST  | /image-management/v1/images/upload                    | identifier: file identification  <br/>part: chunk file.part <br/>priority: Storage priority | ok                                                           | This API only accepts the upload of one file chunk, which is stored in the system according to the identifier |
| cancel upload chunk | DELETE | /image-management/v1/images/upload                    | identifier: file identification  <br/>priority: Storage priority | 取消成功：cancel success/ <br/>取消失败：error               | Canceling will delete the uploaded chunk file.               |
| merge chunk         |  POST  | /image-management/v1/images/merge                     | identifier: file identification  <br/>userId: the ID of user<br/>filename: name for merged file<br/>priority: Storage priority | {<br/>imageId:"string"<br/>fileName:"string"<br/>uploadTime:"string"<br/>userId:"string"<br/>storageMedium:"string"<br/>msg:"string"<br/>checkStatus:int<br/>slimStatus:int<br/>} | slimStatus:[0,1,2,3]stands for Not slimmed yet / slimming down / success / failure<br/>filename:  fill in with original file name which format can be zip or .qcow2/.iso/.img |

