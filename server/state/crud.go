package state

import (
	"fmt"
	"gorm.io/gorm"
)

func Create[T any](db *gorm.DB, entity *T) error {
	result := db.Create(entity)
	return result.Error
}

func GetByID[T any](db *gorm.DB, id string, entity *T) error {
	result := db.First(entity, "id = ?", id)
	return result.Error
}

func GetByKeyVal[T any, R any](db *gorm.DB, key string, val R, entity *T) error {
	result := db.First(entity, fmt.Sprintf("%s = ?", key), val)
	return result.Error
}

func Update[T any](db *gorm.DB, entity *T) error {
	result := db.Save(entity)
	return result.Error
}

func Delete[T any](db *gorm.DB, id string, entity *T) error {
	result := db.Delete(entity, "id = ?", id)
	return result.Error
}

func List[T any](db *gorm.DB, entities *[]T) error {
	result := db.Find(entities)
	return result.Error
}
