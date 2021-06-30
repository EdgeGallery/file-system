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

// @Title  dbAdpater
// @Description  control database
// @Author  GuoZhen Gao (2021/6/30 10:40)
package dbAdpater

import (
	"fileSystem/util"
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"unsafe"
)

//Pg database
type PgDb struct {
	ormer orm.Ormer
}

// Constructor of PluginAdapter
func (db *PgDb) InitOrmer() (err1 error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic handled:", err)
			err1 = fmt.Errorf("recover panic as %s", err)
		}
	}()
	o := orm.NewOrm()
	err1 = o.Using(util.Default)
	if err1 != nil {
		return err1
	}
	db.ormer = o
	return nil
}

// Insert or update data into controller
func (db *PgDb) InsertOrUpdateData(data interface{}, cols ...string) (err error) {
	_, err = db.ormer.InsertOrUpdate(data, cols...)
	return err
}

// Read data from controller
func (db *PgDb) ReadData(data interface{}, cols ...string) (err error) {
	err = db.ormer.Read(data, cols...)
	return err
}

// Delete data from controller
func (db *PgDb) DeleteData(data interface{}, cols ...string) (err error) {
	_, err = db.ormer.Delete(data, cols...)
	return err
}

// Query count for any given table name
func (db *PgDb) QueryCount(tableName string) (int64, error) {
	num, err := db.ormer.QueryTable(tableName).Count()
	return num, err
}

// Query count based on fieldname and fieldvalue
func (db *PgDb) QueryCountForTable(tableName, fieldName, fieldValue string) (int64, error) {
	num, err := db.ormer.QueryTable(tableName).Filter(fieldName, fieldValue).Count()
	return num, err
}

// return a raw query setter for raw sql string.
func (db *PgDb) QueryTable(tableName string, container interface{}, field string, container1 ...interface{}) (num int64, err error) {

	if field != "" {
		num, err = db.ormer.QueryTable(tableName).Filter(field, container1).All(container)
	} else {
		num, err = db.ormer.QueryTable(tableName).All(container)
	}

	return num, err
}

//return the download path
func (db *PgDb) QueryForDownload(tableName string, container interface{}, imageId string) error {
	qs := db.ormer.QueryTable(tableName)
	return qs.Filter("image_id__exact", imageId).One(&container)

}

// Load Related
func (db *PgDb) LoadRelated(md interface{}, name string) (int64, error) {
	num, err := db.ormer.LoadRelated(md, name)
	return num, err
}

func (db *PgDb) InitDatabase() error {
	dbUser := util.GetDbUser()
	dbPwd := []byte(os.Getenv("POSTGRES_PASSWORD"))
	dbName := util.GetDbName()
	dbHost := util.GetDbHost()
	dbPort := util.GetDbPort()
	dbSslMode := util.SslMode
	//dbSslRootCert := DB_SSL_ROOT_CER

	dbPwdStr := string(dbPwd)
/*	util.ClearByteArray(dbPwd)
	dbParamsAreValid, validateDbParamsErr := util.ValidateDbParams(dbPwdStr)
	if validateDbParamsErr != nil || !dbParamsAreValid {
		return errors.New("failed to validate db parameters")
	}*/

	// PostgreSQL configuration
	registerDriverErr := orm.RegisterDriver(util.DriverName, orm.DRPostgres) // register driver
	if registerDriverErr != nil {
		log.Error("Failed to register driver")
		return registerDriverErr
	}

	//get parameters from env
	var b strings.Builder
	fmt.Fprintf(&b, "user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", dbUser, dbPwdStr,
		dbName, dbHost, dbPort, dbSslMode)
	bStr := b.String()

	//registerDataBaseErr := orm.RegisterDataBase("default", "postgres", "user=postgres password=123456 dbname=postgres host=127.0.0.1 port=5432 sslmode=disable")
	registerDataBaseErr := orm.RegisterDataBase(util.Default, util.DriverName, bStr)

	//clear bStr
	bKey1 := *(*[]byte)(unsafe.Pointer(&bStr))
	util.ClearByteArray(bKey1)

	bKey := *(*[]byte)(unsafe.Pointer(&dbPwdStr))
	util.ClearByteArray(bKey)

	if registerDataBaseErr != nil {
		log.Error("Failed to register database")
		return registerDataBaseErr
	}

	// Auto build table
	errRunSyncdb := orm.RunSyncdb(util.Default, false, true)
	if errRunSyncdb != nil {
		log.Error("Failed to sync database.")
		return errRunSyncdb
	}

	err := db.InitOrmer()
	if err != nil {
		log.Error("Failed to init ormer")
		return err
	}

	return nil
}
