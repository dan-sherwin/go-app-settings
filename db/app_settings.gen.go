package db

import (
	"context"
	"github.com/dan-sherwin/go-app-settings/db/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"
)

func newAppSetting(db *gorm.DB, opts ...gen.DOOption) appSetting {
	_appSetting := appSetting{}

	_appSetting.appSettingDo.UseDB(db, opts...)
	_appSetting.appSettingDo.UseModel(&models.AppSetting{})

	tableName := _appSetting.appSettingDo.TableName()
	_appSetting.ALL = field.NewAsterisk(tableName)
	_appSetting.Key = field.NewString(tableName, "key")
	_appSetting.Value = field.NewString(tableName, "value")

	_appSetting.fillFieldMap()

	return _appSetting
}

type appSetting struct {
	appSettingDo

	ALL   field.Asterisk
	Key   field.String
	Value field.String

	fieldMap map[string]field.Expr
}

func (a appSetting) Table(newTableName string) *appSetting {
	a.appSettingDo.UseTable(newTableName)
	return a.updateTableName(newTableName)
}

func (a appSetting) As(alias string) *appSetting {
	a.appSettingDo.DO = *(a.appSettingDo.As(alias).(*gen.DO))
	return a.updateTableName(alias)
}

func (a *appSetting) updateTableName(table string) *appSetting {
	a.ALL = field.NewAsterisk(table)
	a.Key = field.NewString(table, "key")
	a.Value = field.NewString(table, "value")

	a.fillFieldMap()

	return a
}

func (a *appSetting) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := a.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (a *appSetting) fillFieldMap() {
	a.fieldMap = make(map[string]field.Expr, 2)
	a.fieldMap["key"] = a.Key
	a.fieldMap["value"] = a.Value
}

func (a appSetting) clone(db *gorm.DB) appSetting {
	a.appSettingDo.ReplaceConnPool(db.Statement.ConnPool)
	return a
}

func (a appSetting) replaceDB(db *gorm.DB) appSetting {
	a.appSettingDo.ReplaceDB(db)
	return a
}

type appSettingDo struct{ gen.DO }

type IAppSettingDo interface {
	gen.SubQuery
	Debug() IAppSettingDo
	WithContext(ctx context.Context) IAppSettingDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IAppSettingDo
	WriteDB() IAppSettingDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IAppSettingDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IAppSettingDo
	Not(conds ...gen.Condition) IAppSettingDo
	Or(conds ...gen.Condition) IAppSettingDo
	Select(conds ...field.Expr) IAppSettingDo
	Where(conds ...gen.Condition) IAppSettingDo
	Order(conds ...field.Expr) IAppSettingDo
	Distinct(cols ...field.Expr) IAppSettingDo
	Omit(cols ...field.Expr) IAppSettingDo
	Join(table schema.Tabler, on ...field.Expr) IAppSettingDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IAppSettingDo
	RightJoin(table schema.Tabler, on ...field.Expr) IAppSettingDo
	Group(cols ...field.Expr) IAppSettingDo
	Having(conds ...gen.Condition) IAppSettingDo
	Limit(limit int) IAppSettingDo
	Offset(offset int) IAppSettingDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IAppSettingDo
	Unscoped() IAppSettingDo
	Create(values ...*models.AppSetting) error
	CreateInBatches(values []*models.AppSetting, batchSize int) error
	Save(values ...*models.AppSetting) error
	First() (*models.AppSetting, error)
	Take() (*models.AppSetting, error)
	Last() (*models.AppSetting, error)
	Find() ([]*models.AppSetting, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.AppSetting, err error)
	FindInBatches(result *[]*models.AppSetting, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*models.AppSetting) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IAppSettingDo
	Assign(attrs ...field.AssignExpr) IAppSettingDo
	Joins(fields ...field.RelationField) IAppSettingDo
	Preload(fields ...field.RelationField) IAppSettingDo
	FirstOrInit() (*models.AppSetting, error)
	FirstOrCreate() (*models.AppSetting, error)
	FindByPage(offset int, limit int) (result []*models.AppSetting, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IAppSettingDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (a appSettingDo) Debug() IAppSettingDo {
	return a.withDO(a.DO.Debug())
}

func (a appSettingDo) WithContext(ctx context.Context) IAppSettingDo {
	return a.withDO(a.DO.WithContext(ctx))
}

func (a appSettingDo) ReadDB() IAppSettingDo {
	return a.Clauses(dbresolver.Read)
}

func (a appSettingDo) WriteDB() IAppSettingDo {
	return a.Clauses(dbresolver.Write)
}

func (a appSettingDo) Session(config *gorm.Session) IAppSettingDo {
	return a.withDO(a.DO.Session(config))
}

func (a appSettingDo) Clauses(conds ...clause.Expression) IAppSettingDo {
	return a.withDO(a.DO.Clauses(conds...))
}

func (a appSettingDo) Returning(value interface{}, columns ...string) IAppSettingDo {
	return a.withDO(a.DO.Returning(value, columns...))
}

func (a appSettingDo) Not(conds ...gen.Condition) IAppSettingDo {
	return a.withDO(a.DO.Not(conds...))
}

func (a appSettingDo) Or(conds ...gen.Condition) IAppSettingDo {
	return a.withDO(a.DO.Or(conds...))
}

func (a appSettingDo) Select(conds ...field.Expr) IAppSettingDo {
	return a.withDO(a.DO.Select(conds...))
}

func (a appSettingDo) Where(conds ...gen.Condition) IAppSettingDo {
	return a.withDO(a.DO.Where(conds...))
}

func (a appSettingDo) Order(conds ...field.Expr) IAppSettingDo {
	return a.withDO(a.DO.Order(conds...))
}

func (a appSettingDo) Distinct(cols ...field.Expr) IAppSettingDo {
	return a.withDO(a.DO.Distinct(cols...))
}

func (a appSettingDo) Omit(cols ...field.Expr) IAppSettingDo {
	return a.withDO(a.DO.Omit(cols...))
}

func (a appSettingDo) Join(table schema.Tabler, on ...field.Expr) IAppSettingDo {
	return a.withDO(a.DO.Join(table, on...))
}

func (a appSettingDo) LeftJoin(table schema.Tabler, on ...field.Expr) IAppSettingDo {
	return a.withDO(a.DO.LeftJoin(table, on...))
}

func (a appSettingDo) RightJoin(table schema.Tabler, on ...field.Expr) IAppSettingDo {
	return a.withDO(a.DO.RightJoin(table, on...))
}

func (a appSettingDo) Group(cols ...field.Expr) IAppSettingDo {
	return a.withDO(a.DO.Group(cols...))
}

func (a appSettingDo) Having(conds ...gen.Condition) IAppSettingDo {
	return a.withDO(a.DO.Having(conds...))
}

func (a appSettingDo) Limit(limit int) IAppSettingDo {
	return a.withDO(a.DO.Limit(limit))
}

func (a appSettingDo) Offset(offset int) IAppSettingDo {
	return a.withDO(a.DO.Offset(offset))
}

func (a appSettingDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IAppSettingDo {
	return a.withDO(a.DO.Scopes(funcs...))
}

func (a appSettingDo) Unscoped() IAppSettingDo {
	return a.withDO(a.DO.Unscoped())
}

func (a appSettingDo) Create(values ...*models.AppSetting) error {
	if len(values) == 0 {
		return nil
	}
	return a.DO.Create(values)
}

func (a appSettingDo) CreateInBatches(values []*models.AppSetting, batchSize int) error {
	return a.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (a appSettingDo) Save(values ...*models.AppSetting) error {
	if len(values) == 0 {
		return nil
	}
	return a.DO.Save(values)
}

func (a appSettingDo) First() (*models.AppSetting, error) {
	if result, err := a.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.AppSetting), nil
	}
}

func (a appSettingDo) Take() (*models.AppSetting, error) {
	if result, err := a.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.AppSetting), nil
	}
}

func (a appSettingDo) Last() (*models.AppSetting, error) {
	if result, err := a.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.AppSetting), nil
	}
}

func (a appSettingDo) Find() ([]*models.AppSetting, error) {
	result, err := a.DO.Find()
	return result.([]*models.AppSetting), err
}

func (a appSettingDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.AppSetting, err error) {
	buf := make([]*models.AppSetting, 0, batchSize)
	err = a.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (a appSettingDo) FindInBatches(result *[]*models.AppSetting, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return a.DO.FindInBatches(result, batchSize, fc)
}

func (a appSettingDo) Attrs(attrs ...field.AssignExpr) IAppSettingDo {
	return a.withDO(a.DO.Attrs(attrs...))
}

func (a appSettingDo) Assign(attrs ...field.AssignExpr) IAppSettingDo {
	return a.withDO(a.DO.Assign(attrs...))
}

func (a appSettingDo) Joins(fields ...field.RelationField) IAppSettingDo {
	for _, _f := range fields {
		a = *a.withDO(a.DO.Joins(_f))
	}
	return &a
}

func (a appSettingDo) Preload(fields ...field.RelationField) IAppSettingDo {
	for _, _f := range fields {
		a = *a.withDO(a.DO.Preload(_f))
	}
	return &a
}

func (a appSettingDo) FirstOrInit() (*models.AppSetting, error) {
	if result, err := a.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.AppSetting), nil
	}
}

func (a appSettingDo) FirstOrCreate() (*models.AppSetting, error) {
	if result, err := a.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.AppSetting), nil
	}
}

func (a appSettingDo) FindByPage(offset int, limit int) (result []*models.AppSetting, count int64, err error) {
	result, err = a.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = a.Offset(-1).Limit(-1).Count()
	return
}

func (a appSettingDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = a.Count()
	if err != nil {
		return
	}

	err = a.Offset(offset).Limit(limit).Scan(result)
	return
}

func (a appSettingDo) Scan(result interface{}) (err error) {
	return a.DO.Scan(result)
}

func (a appSettingDo) Delete(models ...*models.AppSetting) (result gen.ResultInfo, err error) {
	return a.DO.Delete(models)
}

func (a *appSettingDo) withDO(do gen.Dao) *appSettingDo {
	a.DO = *do.(*gen.DO)
	return a
}
