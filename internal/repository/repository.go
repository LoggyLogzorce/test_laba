package repository

import (
	"database/sql"
	"masha_laba_3/internal/models"
)

type Repository interface {
	GetUserByLogin(login string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	CreateUser(u *models.User) error
	GetUsersAll() ([]models.User, error)
	GetTests() ([]models.Test, error)
	CreateTest(t *models.Test) error
	UpdateTest(t *models.Test) error
	DeleteTest(id int) error
	GetCategory() ([]models.Category, error)
	GetQuestionsByTest(testID uint) ([]models.Question, error)
	GetTestByID(testID uint) (*models.Test, error)
	GetAnswerOptionsByQuestionID(questionID uint) ([]models.AnswerOption, error)
	VulnerableSearch(title string) ([]models.Test, error)
	Update2FASecret(userID uint, secret string) error
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) Repository {
	return &PostgresRepo{db: db}
}

// Auth
func (r *PostgresRepo) GetUserByLogin(login string) (*models.User, error) {
	u := &models.User{}
	// Внимание: кавычки для кириллицы
	err := r.db.QueryRow(`SELECT "id_пользователя", "Логин", "Хеш_пароля", "Роль", "2FA_секрет" FROM users WHERE "Логин" = $1`, login).
		Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Role, &u.TwoFASecret)
	return u, err
}

func (r *PostgresRepo) GetUserByID(id uint) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(`SELECT "id_пользователя", "Логин", "Email", "Роль", "2FA_секрет" FROM users WHERE "id_пользователя" = $1`, id).
		Scan(&u.ID, &u.Login, &u.Email, &u.Role, &u.TwoFASecret)
	return u, err
}

// User
func (r *PostgresRepo) CreateUser(u *models.User) error {
	_, err := r.db.Exec(`INSERT INTO users ("Логин", "Хеш_пароля", "Фамилия", "Имя", "Отчество", "Email", "Роль") VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		u.Login, u.PasswordHash, u.LastName, u.FirstName, u.MiddleName, u.Email, u.Role)
	return err
}

func (r *PostgresRepo) GetUsersAll() ([]models.User, error) {
	rows, err := r.db.Query(`SELECT "Логин", "Фамилия", "Имя", "Отчество", "Email", "Роль" FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.Login, &u.LastName, &u.FirstName, &u.MiddleName, &u.Email, &u.Role)
		users = append(users, u)
	}
	return users, nil
}

// Tests CRUD
func (r *PostgresRepo) GetTests() ([]models.Test, error) {
	rows, err := r.db.Query(`
		SELECT 
			t."id_теста", 
			t."id_пользователя", 
			t."Название", 
			t."Описание", 
			t."Лимит_времени", 
			t."Проходной_балл", 
			t."Активен",
			c."id_категории",
			c."Название" AS "Название_категории"
		FROM tests t
		LEFT JOIN categories c
		ON t.id_категории = c.id_категории
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []models.Test
	for rows.Next() {
		var t models.Test
		var categoryID sql.NullInt64
		var categoryName sql.NullString

		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.TimeLimit,
			&t.PassingScore,
			&t.IsActive,
			&categoryID,
			&categoryName,
		)
		if err != nil {
			return nil, err
		}

		if categoryID.Valid {
			tmp := uint(categoryID.Int64)
			t.CategoryID = &tmp
		}
		if categoryName.Valid {
			t.CategoryName = categoryName.String
		}

		tests = append(tests, t)
	}

	return tests, nil
}

func (r *PostgresRepo) CreateTest(t *models.Test) error {
	_, err := r.db.Exec(`INSERT INTO tests ("id_пользователя", "id_категории", "Название", "Описание", "Лимит_времени", "Проходной_балл", "Активен") VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		t.UserID, t.CategoryID, t.Title, t.Description, t.TimeLimit, t.PassingScore, t.IsActive)
	return err
}

func (r *PostgresRepo) UpdateTest(t *models.Test) error {
	_, err := r.db.Exec(`UPDATE tests SET "Название"=$2, "id_категории"=$3, "Описание"=$4, "Лимит_времени"=$5, "Проходной_балл"=$6, "Активен"=$7 WHERE "id_теста"=$1`,
		t.ID, t.Title, t.CategoryID, t.Description, t.TimeLimit, t.PassingScore, t.IsActive)
	return err
}

func (r *PostgresRepo) DeleteTest(id int) error {
	_, err := r.db.Exec(`DELETE FROM tests WHERE "id_теста" = $1`, id)
	return err
}

func (r *PostgresRepo) GetTestByID(testID uint) (*models.Test, error) {
	row := r.db.QueryRow(`
		SELECT 
			t."id_теста",
			t."id_пользователя",
			t."Название",
			t."Описание",
			t."Лимит_времени",
			t."Проходной_балл",
			t."Активен",
			t.id_категории,
			c."Название"
		FROM tests t
		LEFT JOIN categories c
		ON t.id_категории = c.id_категории
		WHERE t."id_теста" = $1
	`, testID)

	var t models.Test
	var categoryID sql.NullInt64
	var categoryName sql.NullString

	err := row.Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.TimeLimit,
		&t.PassingScore,
		&t.IsActive,
		&categoryID,
		&categoryName,
	)
	if err != nil {
		return nil, err
	}

	if categoryID.Valid {
		tmp := uint(categoryID.Int64)
		t.CategoryID = &tmp
	}

	if categoryName.Valid {
		t.CategoryName = categoryName.String
	}

	return &t, nil
}

func (r *PostgresRepo) VulnerableSearch(title string) ([]models.Test, error) {
	query := `SELECT "id_теста", "id_пользователя", "Название", "Описание", "Лимит_времени", "Проходной_балл", "Активен" FROM tests WHERE "Название" = '` + title + `'`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tests []models.Test
	for rows.Next() {
		var t models.Test
		rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.TimeLimit, &t.PassingScore, &t.IsActive)
		tests = append(tests, t)
	}
	return tests, nil
}

func (r *PostgresRepo) Update2FASecret(userID uint, secret string) error {
	_, err := r.db.Exec(`UPDATE users SET "2FA_секрет" = $1 WHERE "id_пользователя" = $2`, secret, userID)
	return err
}

func (r *PostgresRepo) GetCategory() ([]models.Category, error) {
	rows, err := r.db.Query(`SELECT "id_категории", "Название", "Описание" FROM categories`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cat []models.Category
	for rows.Next() {
		var c models.Category
		rows.Scan(&c.ID, &c.Name, &c.Description)
		cat = append(cat, c)
	}
	return cat, nil
}

func (r *PostgresRepo) GetQuestionsByTest(testID uint) ([]models.Question, error) {
	rows, err := r.db.Query(`SELECT "id_вопроса", "Текст_вопроса", "Тип_вопроса", "Порядковый_номер", "Макс_балл", "Ссылка_на_картинку" FROM questions WHERE "id_теста" = $1`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var questions []models.Question
	for rows.Next() {
		var q models.Question
		rows.Scan(&q.ID, &q.Text, &q.Type, &q.OrderNumber, &q.MaxScore, &q.ImageURL)
		questions = append(questions, q)
	}
	return questions, nil
}

func (r *PostgresRepo) GetAnswerOptionsByQuestionID(questionID uint) ([]models.AnswerOption, error) {
	rows, err := r.db.Query(`
		SELECT 
			"id_варианта",
			"id_вопроса",
			"Текст_варианта",
			"Верный_ли",
			"Порядковый_номер"
		FROM answer_options
		WHERE id_вопроса = $1
		ORDER BY "Порядковый_номер"
	`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []models.AnswerOption

	for rows.Next() {
		var o models.AnswerOption

		err := rows.Scan(
			&o.ID,
			&o.QuestionID,
			&o.Text,
			&o.IsCorrect,
			&o.OrderNumber,
		)
		if err != nil {
			return nil, err
		}

		options = append(options, o)
	}

	return options, nil
}
