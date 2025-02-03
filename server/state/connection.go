package state

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"main/lib"
)

var database *gorm.DB

func createConnection() {
	dbHost := lib.GetDotEnv("PSQL_HOST")
	dbUser := lib.GetDotEnv("PSQL_USER")
	dbPassword := lib.GetDotEnv("PSQL_PASSWORD")
	dbName := lib.GetDotEnv("PSQL_DB")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 TimeZone=Europe/Warsaw", dbHost, dbUser, dbPassword, dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	database = db
}

func GetConnection() *gorm.DB {
	if database == nil {
		createConnection()
	}
	return database
}
