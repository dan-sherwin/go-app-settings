package models

const TableNameAppSetting = "app_settings"

type AppSetting struct {
	Key         string `gorm:"column:key;type:TEXT;primaryKey" json:"key"`
	Value       string `gorm:"column:value;type:TEXT;not null" json:"value"`
	Description string `gorm:"-" json:"description"`
}

func (*AppSetting) TableName() string {
	return TableNameAppSetting
}
