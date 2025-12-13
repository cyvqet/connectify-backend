package service

import (
	"context"
	"errors"

	"connectify/internal/domain"
	"connectify/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrUserNotFound          = repository.ErrUserNotFound
	ErrInvaildUserOrPassword = errors.New("invalid username or password")
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, user domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return svc.repo.Create(ctx, user)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) error {
	user, err := svc.repo.FindByEmail(ctx, email)
	if err == ErrUserNotFound {
		return ErrInvaildUserOrPassword
	}
	if err != nil {
		return err
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Log error
		return ErrInvaildUserOrPassword
	}
	return nil
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindById(ctx, id)
}

func (svc *UserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// Check if user exists by phone number
	user, err := svc.repo.FindByPhone(ctx, phone)
	if err == nil {
		// User exists, return directly
		return user, nil
	}

	if err != ErrUserNotFound {
		// Other errors, return directly
		return domain.User{}, err
	}

	// User does not exist, create new user
	newUser := domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, newUser)
	if err != nil {
		return domain.User{}, err
	}

	// Query again after creation to get complete information (including ID)
	createdUser, err := svc.repo.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}

	return createdUser, nil

}
