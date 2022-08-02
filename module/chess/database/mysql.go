package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MysqlDatabase mysql database struct
type MysqlDatabase struct {
	UserName string
	Password string
	Host     string
	Port     string
	Database string
	CharSet  string
}

// InitDB init database
func (m MysqlDatabase) InitDB(migrateDst ...interface{}) (*gorm.DB, error) {
	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		m.UserName,
		m.Password,
		m.Host,
		m.Port,
		m.Database,
		m.CharSet)
	db, err := gorm.Open(mysql.Open(args), &gorm.Config{})
	if err != nil {
		return db, err
	}
	db.AutoMigrate(migrateDst...)

	return db, nil
}
