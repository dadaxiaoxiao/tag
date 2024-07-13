package dao

import "gorm.io/gorm"

// InitTables 初始化表
func InitTables(db *gorm.DB) error {
	// 自动迁移模式
	return db.AutoMigrate(
		&Tag{},
		&TagBiz{})
}
