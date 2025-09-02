package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dan-sherwin/go-app-settings/db/models"
	"gorm.io/driver/sqlite"

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
	var err error
	DB, err = gorm.Open(sqlite.Open(fileName), &gorm.Config{})
	if err != nil {
		return err
	}
	sqldb, _ := DB.DB()
	if err := sqldb.Ping(); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	if err := DB.AutoMigrate(&models.AppSetting{}); err != nil {
		return err
	}
	SetDefault(DB)
	return nil
}

func SetDefault(db *gorm.DB, opts ...gen.DOOption) {
	*Q = *Use(db, opts...)
	AppSetting = &Q.AppSetting
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
