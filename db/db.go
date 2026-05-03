package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dan-sherwin/go-app-settings/db/models"
	"github.com/glebarez/sqlite"

	"gorm.io/gorm"

	"gorm.io/gen"

	"gorm.io/plugin/dbresolver"
)

var (
	Q          = new(Query)
	AppSetting *appSetting
	DB         *gorm.DB
)

func DBInit(fileName string) error {
	return DBInitWithTable(fileName, models.TableNameAppSetting)
}

func DBInitWithTable(fileName string, tableName string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(fileName), &gorm.Config{})
	if err != nil {
		return err
	}
	return DBInitWithDB(DB, tableName)
}

func DBInitWithDB(gormDB *gorm.DB, tableName string) error {
	if gormDB == nil {
		return fmt.Errorf("database cannot be nil")
	}
	if tableName == "" {
		tableName = models.TableNameAppSetting
	}
	DB = gormDB
	sqldb, err := DB.DB()
	if err != nil {
		return fmt.Errorf("unable to get database handle: %w", err)
	}
	if err := sqldb.Ping(); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	if err := ensureAppSettingsTable(DB, tableName); err != nil {
		return err
	}
	SetDefaultTable(DB, tableName)
	return nil
}

func ensureAppSettingsTable(gormDB *gorm.DB, tableName string) error {
	migrator := gormDB.Migrator()
	if !migrator.HasTable(tableName) {
		return gormDB.Table(tableName).AutoMigrate(&models.AppSetting{})
	}
	cols, err := migrator.ColumnTypes(tableName)
	if err != nil {
		return fmt.Errorf("inspect %s table: %w", tableName, err)
	}
	hasKey := false
	keyPrimary := true
	hasValue := false
	for _, col := range cols {
		switch col.Name() {
		case "key":
			hasKey = true
			if primary, ok := col.PrimaryKey(); ok {
				keyPrimary = primary
			}
		case "value":
			hasValue = true
		}
	}
	if !hasKey || !hasValue {
		return fmt.Errorf("table %s already exists but is not compatible with app_settings: required columns key and value", tableName)
	}
	if !keyPrimary {
		return fmt.Errorf("table %s already exists but is not compatible with app_settings: key column must be the primary key", tableName)
	}
	return gormDB.Table(tableName).AutoMigrate(&models.AppSetting{})
}

func SetDefault(db *gorm.DB, opts ...gen.DOOption) {
	SetDefaultTable(db, models.TableNameAppSetting, opts...)
}

func SetDefaultTable(db *gorm.DB, tableName string, opts ...gen.DOOption) {
	if tableName == "" {
		tableName = models.TableNameAppSetting
	}
	*Q = *Use(db, opts...)
	AppSetting = Q.AppSetting.Table(tableName)
}

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:         db,
		AppSetting: newAppSetting(db, opts...),
	}
}

type Query struct {
	db *gorm.DB

	AppSetting appSetting
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:         db,
		AppSetting: q.AppSetting.clone(db),
	}
}

func (q *Query) ReadDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Read))
}

func (q *Query) WriteDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Write))
}

func (q *Query) ReplaceDB(db *gorm.DB) *Query {
	return &Query{
		db:         db,
		AppSetting: q.AppSetting.replaceDB(db),
	}
}

type queryCtx struct {
	AppSetting IAppSettingDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		AppSetting: q.AppSetting.WithContext(ctx),
	}
}

func (q *Query) Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(func(tx *gorm.DB) error { return fc(q.clone(tx)) }, opts...)
}

func (q *Query) Begin(opts ...*sql.TxOptions) *QueryTx {
	tx := q.db.Begin(opts...)
	return &QueryTx{Query: q.clone(tx), Error: tx.Error}
}

type QueryTx struct {
	*Query
	Error error
}

func (q *QueryTx) Commit() error {
	return q.db.Commit().Error
}

func (q *QueryTx) Rollback() error {
	return q.db.Rollback().Error
}

func (q *QueryTx) SavePoint(name string) error {
	return q.db.SavePoint(name).Error
}

func (q *QueryTx) RollbackTo(name string) error {
	return q.db.RollbackTo(name).Error
}
