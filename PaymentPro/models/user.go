package models

import (
        "database/sql"
        "fmt"
        "time"
)

type User struct {
        ID              int       `json:"id"`
        TelegramID      int64     `json:"telegram_id"`
        Username        string    `json:"username"`
        FirstName       string    `json:"first_name"`
        LastName        string    `json:"last_name"`
        LanguageCode    string    `json:"language_code"`
        CurrentPlanID   int       `json:"current_plan_id"`
        PlanExpiresAt   *time.Time `json:"plan_expires_at"`
        WebUsername     string    `json:"web_username"`
        WebPassword     string    `json:"web_password"`
        IsWebActive     bool      `json:"is_web_active"`
        CreatedAt       time.Time `json:"created_at"`
        UpdatedAt       time.Time `json:"updated_at"`
}

type UserRepository struct {
        db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
        return &UserRepository{db: db}
}

func (r *UserRepository) GetByTelegramID(telegramID int64) (*User, error) {
        user := &User{}
        query := `
                SELECT id, telegram_id, username, first_name, last_name, language_code, 
                       current_plan_id, plan_expires_at, web_username, web_password, is_web_active, created_at, updated_at
                FROM users WHERE telegram_id = $1
        `
        err := r.db.QueryRow(query, telegramID).Scan(
                &user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName,
                &user.LanguageCode, &user.CurrentPlanID, &user.PlanExpiresAt, &user.WebUsername, &user.WebPassword, &user.IsWebActive, &user.CreatedAt, &user.UpdatedAt,
        )
        if err != nil {
                return nil, err
        }
        return user, nil
}

func (r *UserRepository) CreateOrUpdate(user *User) error {
        query := `
                INSERT INTO users (telegram_id, username, first_name, last_name, language_code, current_plan_id, plan_expires_at)
                VALUES ($1, $2, $3, $4, $5, 1, NOW() + INTERVAL '30 days')
                ON CONFLICT (telegram_id) DO UPDATE SET
                        username = EXCLUDED.username,
                        first_name = EXCLUDED.first_name,
                        last_name = EXCLUDED.last_name,
                        language_code = EXCLUDED.language_code,
                        updated_at = CURRENT_TIMESTAMP
                RETURNING id
        `
        return r.db.QueryRow(query, user.TelegramID, user.Username, user.FirstName, user.LastName, user.LanguageCode).Scan(&user.ID)
}

func (r *UserRepository) UpdateSubscription(userID int, planID int, expiresAt *time.Time) error {
        query := `UPDATE users SET current_plan_id = $1, plan_expires_at = $2 WHERE id = $3`
        _, err := r.db.Exec(query, planID, expiresAt, userID)
        return err
}

func (r *UserRepository) GetExpiredSubscriptions() ([]User, error) {
        query := `
                SELECT id, telegram_id, username, first_name, last_name, language_code, 
                       current_plan_id, plan_expires_at, created_at, updated_at
                FROM users 
                WHERE plan_expires_at < CURRENT_TIMESTAMP AND current_plan_id > 1
        `
        rows, err := r.db.Query(query)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        var users []User
        for rows.Next() {
                var user User
                err := rows.Scan(
                        &user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName,
                        &user.LanguageCode, &user.CurrentPlanID, &user.PlanExpiresAt, &user.CreatedAt, &user.UpdatedAt,
                )
                if err != nil {
                        return nil, err
                }
                users = append(users, user)
        }
        return users, nil
}

func (r *UserRepository) GetExpiringSoon(days int) ([]User, error) {
        query := `
                SELECT id, telegram_id, username, first_name, last_name, language_code, 
                       current_plan_id, plan_expires_at, created_at, updated_at
                FROM users 
                WHERE plan_expires_at BETWEEN CURRENT_TIMESTAMP AND CURRENT_TIMESTAMP + INTERVAL '%d days'
                AND current_plan_id > 1
        `
        rows, err := r.db.Query(fmt.Sprintf(query, days))
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        var users []User
        for rows.Next() {
                var user User
                err := rows.Scan(
                        &user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName,
                        &user.LanguageCode, &user.CurrentPlanID, &user.PlanExpiresAt, &user.CreatedAt, &user.UpdatedAt,
                )
                if err != nil {
                        return nil, err
                }
                users = append(users, user)
        }
        return users, nil
}

func (r *UserRepository) CreateWebAccount(userID int, username, password string) error {
        query := `
                UPDATE users 
                SET web_username = $1, web_password = $2, is_web_active = true, updated_at = CURRENT_TIMESTAMP
                WHERE id = $3
        `
        _, err := r.db.Exec(query, username, password, userID)
        return err
}

func (r *UserRepository) GetByWebCredentials(username, password string) (*User, error) {
        user := &User{}
        query := `
                SELECT id, telegram_id, username, first_name, last_name, language_code, 
                       current_plan_id, plan_expires_at, web_username, web_password, is_web_active, created_at, updated_at
                FROM users 
                WHERE web_username = $1 AND web_password = $2 AND is_web_active = true
        `
        var lastName sql.NullString
        err := r.db.QueryRow(query, username, password).Scan(
                &user.ID, &user.TelegramID, &user.Username, &user.FirstName, &lastName,
                &user.LanguageCode, &user.CurrentPlanID, &user.PlanExpiresAt, &user.WebUsername, &user.WebPassword, &user.IsWebActive, &user.CreatedAt, &user.UpdatedAt,
        )
        if lastName.Valid {
                user.LastName = lastName.String
        }
        if err != nil {
                return nil, err
        }
        return user, nil
}

func (r *UserRepository) GetByID(id int) (*User, error) {
        user := &User{}
        query := `
                SELECT id, telegram_id, username, first_name, last_name, language_code, 
                       current_plan_id, plan_expires_at, web_username, web_password, is_web_active, created_at, updated_at
                FROM users 
                WHERE id = $1
        `
        var lastName sql.NullString
        err := r.db.QueryRow(query, id).Scan(
                &user.ID, &user.TelegramID, &user.Username, &user.FirstName, &lastName,
                &user.LanguageCode, &user.CurrentPlanID, &user.PlanExpiresAt, &user.WebUsername, &user.WebPassword, &user.IsWebActive, &user.CreatedAt, &user.UpdatedAt,
        )
        if lastName.Valid {
                user.LastName = lastName.String
        }
        if err != nil {
                return nil, err
        }
        return user, nil
}



func (r *UserRepository) UpdateWebPassword(userID int, newPassword string) error {
        query := `UPDATE users SET web_password = $1 WHERE id = $2`
        _, err := r.db.Exec(query, newPassword, userID)
        return err
}

func (r *UserRepository) UpdateWebUsername(userID int, newUsername string) error {
        query := `UPDATE users SET web_username = $1 WHERE id = $2`
        _, err := r.db.Exec(query, newUsername, userID)
        return err
}

func (r *UserRepository) UpdateWebCredentials(userID int, username, password string) error {
        query := `
                UPDATE users SET 
                        web_username = $1, 
                        web_password = $2, 
                        is_web_active = TRUE,
                        updated_at = CURRENT_TIMESTAMP
                WHERE id = $3
        `
        _, err := r.db.Exec(query, username, password, userID)
        return err
}
