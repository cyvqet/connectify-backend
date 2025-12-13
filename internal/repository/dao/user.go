package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("email conflict")
	ErrUserNotFound       = errors.New("user not found")
)

type User struct {
	Id        int64          `gorm:"primaryKey,autoIncrement"`
	Email     sql.NullString `gorm:"unique"`
	Phone     sql.NullString `gorm:"unique"`
	Password  string
	CreatedAt int64
	UpdatedAt int64
}

type UserDao interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
}

type gormUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &gormUserDao{
		db: db,
	}
}

func (dao *gormUserDao) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := dao.db.WithContext(ctx).Create(&user).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == 1062 {
			return ErrUserDuplicateEmail // email conflict
		}
	}

	return err
}

func (dao *gormUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, ErrUserNotFound
	}
	return user, err
}

func (dao *gormUserDao) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, ErrUserNotFound
	}
	return user, err
}

func (dao *gormUserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("phone=?", phone).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, ErrUserNotFound
	}
	return user, nil
}
