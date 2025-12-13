package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"connectify/internal/domain"
	"connectify/internal/repository/cache"
	"connectify/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
}

type userRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDao, cache cache.UserCache) UserRepository {
	return &userRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *userRepository) Create(ctx context.Context, user domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Phone:    sql.NullString{String: user.Phone, Valid: user.Phone != ""},
		Email:    sql.NullString{String: user.Email, Valid: user.Email != ""},
		Password: user.Password,
	})
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return r.entityToDomain(user), nil
}

func (r *userRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// Get from cache first
	user, err := r.cache.Get(ctx, id)
	if err == nil {
		fmt.Println("cache hit, get from cache")
		// cache hit
		return user, nil
	}

	// Only query the database if the cache is not hit, other errors are returned directly
	// If you want to downgrade (also query the database if the cache is out of error), you can remove this judgment
	if err != cache.ErrKeyNotExist {
		return domain.User{}, err
	}

	fmt.Println("cache miss, get from database")

	// Cache miss, get from database
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	user = r.entityToDomain(u)

	// Write back to cache, only log if failed
	if err := r.cache.Set(ctx, user); err != nil {
		log.Printf("WARN: write back to cache failed, userId: %d, err: %v", user.Id, err)
	}

	return user, nil
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err == dao.ErrUserNotFound {
		return domain.User{}, ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

// 内部辅助方法，改为小写私有
func (r *userRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Phone:    u.Phone.String,
		Email:    u.Email.String,
		Password: u.Password,
	}
}

func (r *userRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id:       u.Id,
		Phone:    sql.NullString{String: u.Phone, Valid: u.Phone != ""},
		Email:    sql.NullString{String: u.Email, Valid: u.Email != ""},
		Password: u.Password,
	}
}
