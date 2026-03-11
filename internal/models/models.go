package models

import "time"

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleTester  UserRole = "tester"
	RoleStudent UserRole = "student"
)

type User struct {
	ID               uint      `json:"id_пользователя"`
	Login            string    `json:"логин"`
	PasswordHash     string    `json:"-"` // Пароль не отдаем в JSON
	LastName         string    `json:"фамилия"`
	FirstName        string    `json:"имя"`
	MiddleName       *string   `json:"отчество,omitempty"`
	Email            string    `json:"email"`
	Role             UserRole  `json:"роль"`
	RegistrationDate time.Time `json:"дата_регистрации"`
	TwoFASecret      *string   `json:"2fa_секрет,omitempty"`
}

type Category struct {
	ID          uint    `json:"id_категории"`
	Name        string  `json:"название"`
	Description *string `json:"описание,omitempty"`
}

type Test struct {
	ID           uint    `json:"id_теста"`
	UserID       uint    `json:"id_пользователя"`
	CategoryID   *uint   `json:"id_категории,omitempty"`
	CategoryName string  `json:"название_категории"`
	Title        string  `json:"название"`
	Description  *string `json:"описание,omitempty"`
	TimeLimit    int     `json:"лимит_времени"`
	PassingScore int     `json:"проходной_балл"`
	IsActive     bool    `json:"активен"`
	ImageURL     *string `json:"ссылка_на_картинку,omitempty"`
}

type QuestionType string

const (
	QuestionSingle   QuestionType = "single"
	QuestionMultiple QuestionType = "multiple"
	QuestionText     QuestionType = "text"
)

type Question struct {
	ID          uint         `json:"id_вопроса"`
	TestID      uint         `json:"id_теста"`
	Text        string       `json:"текст_вопроса"`
	Type        QuestionType `json:"тип_вопроса"`
	OrderNumber int          `json:"порядковый_номер"`
	MaxScore    int          `json:"макс_балл"`
	ImageURL    *string      `json:"ссылка_на_картинку,omitempty"`
}

type AnswerOption struct {
	ID          uint   `json:"id_варианта"`
	QuestionID  uint   `json:"id_вопроса"`
	Text        string `json:"текст_варианта"`
	IsCorrect   bool   `json:"верный_ли"`
	OrderNumber int    `json:"порядковый_номер"`
}

type AttemptStatus string

const (
	StatusInProgress AttemptStatus = "in_progress"
	StatusCompleted  AttemptStatus = "completed"
)

type TestAttempt struct {
	ID         uint          `json:"id_попытки"`
	UserID     uint          `json:"id_пользователя"`
	StartTime  time.Time     `json:"время_начала"`
	EndTime    *time.Time    `json:"время_окончания,omitempty"`
	TotalScore *int          `json:"итоговый_балл,omitempty"`
	Status     AttemptStatus `json:"статус"`
	IsPassed   bool          `json:"сдан_ли"`
}

type AttemptAnswer struct {
	ID          uint    `json:"id_записи"`
	AttemptID   uint    `json:"id_попытки"`
	QuestionID  uint    `json:"id_вопроса"`
	AnswerText  *string `json:"текст_ответа,omitempty"`
	IsCorrect   *bool   `json:"верный_ли,omitempty"`
	ScoreGained int     `json:"полученный_балл"`
}
