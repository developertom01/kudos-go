package data

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	connection gorm.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	connection, err :=  gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return &Database{
		connection: *connection,
	}, nil
}
