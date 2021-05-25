package dbAdpater

// Database API's
type Database interface {
	InitDatabase() error
	InsertOrUpdateData(data interface{}, cols ...string) (err error)
	ReadData(data interface{}, cols ...string) (err error)
	DeleteData(data interface{}, cols ...string) (err error)
	QueryCount(tableName string) (int64, error)
	QueryCountForTable(tableName, fieldName, fieldValue string) (int64, error)
	QueryTable(query string, container interface{}, field string, container1 ...interface{}) (num int64, err error)
	QueryForDownload(tableName string, container interface{}, imageId string) error
	LoadRelated(md interface{}, name string) (int64, error)
}
