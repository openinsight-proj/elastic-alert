package utils

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	} else {
		return s.IsDir()
	}
}

func MD5(raw string) string {
	bs := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", bs)
}

// ESIdsLenLimit: VM label length limit is 16384 bytes.
// So, we limit 30 * 512 bytes(elasticsearch _doc id lenth limit) = 15360.
func ESIdsLenLimit(ids []string) []string {
	if len(ids) <= 30 {
		return ids
	} else {
		return ids[:30]
	}
}

func PrepareTestDB(sqlDB *sql.DB) (db *gorm.DB) {
	db, _ = gorm.Open(mysql.New(mysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      sqlDB,
	}), &gorm.Config{})
	return
}
