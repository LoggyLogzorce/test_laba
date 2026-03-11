package services

import (
	"errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"masha_laba_3/internal/models"
	"masha_laba_3/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrInvalid2FACode     = errors.New("неверный 2FA код")
	ErrUserNotFound       = errors.New("пользователь не найден")
)

type Services interface {
	GetAllUsers() ([]models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserByLogin(login string) (*models.User, error)
	GetTests() ([]models.Test, error)
	CreateTest(t *models.Test) error
	UpdateTest(t *models.Test) error
	DeleteTest(id int) error
	GetCategories() ([]models.Category, error)
	GetQuestionsByTest(testID uint) ([]models.Question, error)
	GetTestByID(testID uint) (*models.Test, error)
	GetAnswerOptionsByQuestionID(questionID uint) ([]models.AnswerOption, error)
	Login(login, password string) (*Auth, error)
	Register(data *models.User) error
	Verify2FA(secret, code string) error
	Generate2FASecret(login, email string) (*otp.Key, string, error)
	Enable2FA(userID uint, secret string) error
	SQLi(title string) ([]models.Test, error)
}

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Services {
	return &Service{repo: repo}
}

type Auth struct {
	Ok     bool
	UserID uint
	Role   models.UserRole
	TwoFA  string
}

// Login проверяет учётные данные
func (s *Service) Login(login, password string) (*Auth, error) {
	user, err := s.repo.GetUserByLogin(login)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	result := &Auth{
		Ok:     true,
		UserID: user.ID,
		Role:   user.Role,
	}

	if user.TwoFASecret == nil {
		result.TwoFA = ""
	} else {
		result.TwoFA = *user.TwoFASecret
	}

	return result, nil
}

func (s *Service) Register(data *models.User) error {
	// Хеширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(data.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	data.PasswordHash = string(hash)

	return s.repo.CreateUser(data)
}

// Verify2FA проверяет TOTP код
func (s *Service) Verify2FA(secret, code string) error {
	if secret == "" {
		return ErrInvalid2FACode
	}

	// Проверка TOTP (окно 30 сек, ±1 период)
	ok := totp.Validate(code, secret)
	if !ok {
		return ErrInvalid2FACode
	}
	return nil
}

// Generate2FASecret генерирует секрет для 2FA
func (s *Service) Generate2FASecret(login, email string) (*otp.Key, string, error) {
	// Используем и login, и email для более понятного отображения в приложении
	accountName := login
	if email != "" {
		accountName = login + " (" + email + ")"
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "TestApp",
		AccountName: accountName,
	})
	if err != nil {
		return nil, "", err
	}

	return key, key.Secret(), nil
}

// Enable2FA включает 2FA для пользователя
func (s *Service) Enable2FA(userID uint, secret string) error {
	return s.repo.Update2FASecret(userID, secret)
}

func (s *Service) GetAllUsers() ([]models.User, error) {
	return s.repo.GetUsersAll()
}

func (s *Service) GetUserByID(id uint) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *Service) GetUserByLogin(login string) (*models.User, error) {
	return s.repo.GetUserByLogin(login)
}

func (s *Service) GetTests() ([]models.Test, error) {
	return s.repo.GetTests()
}

func (s *Service) CreateTest(t *models.Test) error {
	return s.repo.CreateTest(t)
}

func (s *Service) UpdateTest(t *models.Test) error {
	return s.repo.UpdateTest(t)
}

func (s *Service) DeleteTest(id int) error {
	return s.repo.DeleteTest(id)
}

func (s *Service) GetCategories() ([]models.Category, error) {
	return s.repo.GetCategory()
}

func (s *Service) SQLi(title string) ([]models.Test, error) {
	tests, err := s.repo.VulnerableSearch(title)
	if err != nil {
		return nil, err
	}
	return tests, nil
}

func (s *Service) GetQuestionsByTest(testID uint) ([]models.Question, error) {
	return s.repo.GetQuestionsByTest(testID)
}

func (s *Service) GetTestByID(testID uint) (*models.Test, error) {
	return s.repo.GetTestByID(testID)
}

func (s *Service) GetAnswerOptionsByQuestionID(questionID uint) ([]models.AnswerOption, error) {
	return s.repo.GetAnswerOptionsByQuestionID(questionID)
}
