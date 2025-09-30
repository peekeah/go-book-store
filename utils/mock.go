package utils

import (
	"fmt"
	"log"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/peekeah/book-store/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBMock struct {
	Db   *gorm.DB
	Mock sqlmock.Sqlmock
}

var DB DBMock

func GetDBMock() (*gorm.DB, sqlmock.Sqlmock) {
	if DB != (DBMock{}) {
		return DB.Db, DB.Mock
	}

	dbConfig := config.GetConfig().DB
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.Port,
	)

	mockDb, mock, err := sqlmock.NewWithDSN(dsn)
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening gorm database", err)
	}

	DB.Db = db
	DB.Mock = mock

	return db, mock
}
