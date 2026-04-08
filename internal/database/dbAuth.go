package database

import (
	"gorm.io/gorm"
)


var DbAuth *gorm.DB

func OpenAuth() {

	DbAuth = DbWebkita
}