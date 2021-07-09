#EdgeGallery file-system



##Overview

   file-system provides developers with file management services for uploading, downloading, querying, and deleting



##Detailed introduction

*upload

  The upload file can be the image file itself. If the image file is too large, it is recommended to compress it into a zip file and upload it. When uploading a zip, the .qocw2 file needs to be compressed in the folder

*download

  Different formats can be selected when downloading files according to imageId, which can be controlled by QUERY parameters

*query

 Query file details based on imageId

*delete

  Delete stored files and data information according to imageId



##API Defination

|          | Method | URL                                                   | form-data参数                                                | 相应结构                                                     | 接口实现说明                                                 |
| -------- | :----: | ----------------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| upload   |  POST  | /image-management/v1/images                           | userId: the ID of user<br/>file: chosen file<br/>priority: Storage priority | {imageId:"string"<br/>file:"string"<br/>uploadTime:"string"<br/>userId:"string"<br/>storageMedium:"string"} | Optional upload image file format：.zip/.qcow2/.img/.iso, ；***priority*** generally select 0；When uploading the zip, the upper layer of the image file should be wrapped with a layer of folders |
| download |  GET   | /image-management/v1/images/{imageId}/action/download | None                                                         | file                                                         | The format of the download image file is optional. When the query is /?isZip=true, the download format is .zip; when the query is not included, the image file itself is downloaded |
| query    |  GET   | /image-management/v1/images/{imageId}                 | None                                                         | {imageId:"string"<br/>file:"string"<br/>uploadTime:"string"<br/>userId:"string"<br/>storageMedium:"string"} | Query file details based on imageId                          |
| delete   | DELETE | /image-management/v1/images/{imageId}                 | None                                                         | 删除成功: delete success/<br/>删除失败: error                | Delete local files based on imageId                          |