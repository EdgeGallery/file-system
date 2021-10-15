# EdgeGallery file-system



## 概述

​    file-system为developer提供上传、下载、查询、删除的文件管理服务



## 详细介绍

- 上传

  上传文件可选镜像文件本身，若镜像文件过大，建议压缩成zip文件上传，上传zip时，.qocw2文件需放在文件夹里压缩

- 下载

  根据imageId下载文件时可选不同格式，通过query参数来控制

- 查询

  根据imageId查询文件详情

- 删除

  根据imageId删除存储的文件及数据信息



## 接口定义

|              | Method | URL                                                   | form-data参数                                       | 相应结构                                                     | 接口实现说明                                                 |
| ------------ | :----: | ----------------------------------------------------- | --------------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| 上传镜像文件 |  POST  | /image-management/v1/images                           | userId:用户ID<br/>file:文件<br/>priority:存储优先级 | {imageId:"string"<br/>file:"string"<br/>uploadTime:"string"<br/>userId:"string"<br/>storageMedium:"string"} | 上传镜像文件格式可选：.zip/.qcow2/.img/.iso, ；priority一般选0；上传zip时，镜像文件上层应包一层文件夹 |
| 下载镜像文件 |  GET   | /image-management/v1/images/{imageId}/action/download | 无                                                  | file                                                         | 下载镜像文件格式可选，query为/?isZip=true时下载格式为.zip；不带query时下载镜像文件本身 |
| 查询虚机镜像 |  GET   | /image-management/v1/images/{imageId}                 | 无                                                  | {imageId:"string"<br/>file:"string"<br/>uploadTime:"string"<br/>userId:"string"<br/>storageMedium:"string"} | 根据imageId查询文件详情                                      |
| 删除虚机镜像 | DELETE | /image-management/v1/images/{imageId}                 | 无                                                  | 删除成功: delete success/<br/>删除失败: error                | 根据imageId删除本地文件                                      |