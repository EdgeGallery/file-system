package dbAdpater

import (
	"errors"
	"fileSystem/util"
)


// Init Db adapter
func GetDbAdapter() (Database, error) {
	dbAdapter := util.GetAppConfig("dbAdapter")
	dbAdapter = "pgDb"
	switch dbAdapter {
	case "pgDb":
		db := &PgDb{}
		err := db.InitDatabase()
		if err != nil {
			return nil, errors.New("failed to register database")
		}
		return db, nil
	default:
		return nil, errors.New("no database is found")
	}
}