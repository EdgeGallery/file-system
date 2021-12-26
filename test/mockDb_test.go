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

package test

import (
	"errors"
	"fileSystem/models"
	"fileSystem/util"
)

type MockDb struct {
	imageRecords models.ImageDB
}

func (db *MockDb) InitDatabase() error {
	panic("implement me")
}

func (db *MockDb) InsertOrUpdateData(data interface{}, cols ...string) (err error) {
	if cols[0] == util.DbImageId {
		imageDb, ok := data.(*models.ImageDB)
		if ok {
			db.imageRecords = *imageDb
		}
	}
	return nil
}

func (db *MockDb) ReadData(data interface{}, cols ...string) (err error) {
	if cols[0] == util.DbImageId {
		imageDb, ok := data.(*models.ImageDB)
		if ok {
			readImageData := db.imageRecords
			if (readImageData == models.ImageDB{}) {
				return errors.New("Image DB record not found ")
			}
			//TODO:补全
			imageDb.FileName = readImageData.FileName
			imageDb.SaveFileName = readImageData.SaveFileName
		}
	}
	return nil
}

func (db *MockDb) DeleteData(data interface{}, cols ...string) (err error) {
	if cols[0] == util.DbImageId {
		_, ok := data.(*models.ImageDB)
		if ok {
			readImageData := db.imageRecords
			if (readImageData == models.ImageDB{}) {
				return errors.New("Image DB record not found ")
			}
			db.imageRecords = models.ImageDB{}
		}
	}
	return nil
}

func (db *MockDb) QueryCount(tableName string) (int64, error) {
	return 0, nil
}

func (db *MockDb) QueryCountForTable(tableName, fieldName, fieldValue string) (int64, error) {
	if tableName == "image_d_b" {
		return 1, nil
	}
	return 0, nil
}

func (db *MockDb) QueryTable(tableName string, container interface{}, field string, container1 ...interface{}) (int64, error) {
	if tableName == "image_d_b" {
		return 1, nil
	}
	return 0, nil
}

func (db *MockDb) LoadRelated(md interface{}, name string) (int64, error) {
	return 0, nil
}
