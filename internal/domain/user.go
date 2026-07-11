package domain

import (
	"errors"
	"time"
)

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleLibrarian Role = "librarian"
	RoleReader    Role = "reader"
)

type User struct {
	ID           int        `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         Role       `json:"role"`
	ReaderID     *int       `json:"reader_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

func NewUser(email, passwordHash string, role Role, readerID *int) User {
	now := time.Now()
	return User{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		ReaderID:     readerID,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  nil,
	}
}

func (u *User) Validate() error {
	// Проверка email
	if u.Email == "" {
		return errors.New("email не может быть пустым ")
	}

	if !isValidEmail(u.Email) {
		return errors.New("Не верный формат email")
	}

	if u.PasswordHash == "" {
		return errors.New("Пароль не может быть пустым ")
	}

	if !isValidRole(u.Role) {
		return errors.New("Нету роли")
	}

	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsLibrarian() bool {
	return u.Role == RoleLibrarian || u.Role == RoleAdmin
}

func (u *User) IsReader() bool {
	return u.Role == RoleReader
}

func (u *User) HasReaderProfile() bool {
	return u.ReaderID != nil && *u.ReaderID > 0
}

func (u *User) CanAccessAdminPanel() bool {
	return u.Role == RoleAdmin || u.Role == RoleLibrarian
}

func (u *User) CanManageBooks() bool {
	return u.Role == RoleAdmin || u.Role == RoleLibrarian
}

func (u *User) CanManageReaders() bool {
	return u.Role == RoleAdmin || u.Role == RoleLibrarian
}

func (u *User) CanManageTransactions() bool {
	return u.Role == RoleAdmin || u.Role == RoleLibrarian
}

func (u *User) CanViewReports() bool {
	return u.Role == RoleAdmin || u.Role == RoleLibrarian
}

func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

func (u *User) UpdatePassword(newHash string) {
	u.PasswordHash = newHash
	u.UpdatedAt = time.Now()
}

func (u *User) UpdateEmail(newEmail string) error {
	if newEmail == "" {
		return errors.New("Новый email не может быть пустым")
	}
	if !isValidEmail(newEmail) {
		return errors.New("Не верный формат email")
	}
	u.Email = newEmail
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) UpdateRole(newRole Role) error {
	if !isValidRole(newRole) {
		return errors.New("Нету роли")
	}
	u.Role = newRole
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) LinkReader(readerID int) {
	u.ReaderID = &readerID
	u.UpdatedAt = time.Now()
}

func (u *User) UnlinkReader() {
	u.ReaderID = nil
	u.UpdatedAt = time.Now()
}

func isValidEmail(email string) bool {
	if len(email) < 3 {
		return false
	}

	atIndex := -1
	dotIndex := -1

	for i, ch := range email {
		if ch == '@' {
			atIndex = i
		}
		if ch == '.' && atIndex != -1 {
			dotIndex = i
		}
	}

	return atIndex > 0 && dotIndex > atIndex+1
}

func isValidRole(role Role) bool {
	switch role {
	case RoleAdmin, RoleLibrarian, RoleReader:
		return true
	default:
		return false
	}
}

type UserBuilder struct {
	user User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: User{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func (b *UserBuilder) WithID(id int) *UserBuilder {
	b.user.ID = id
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) WithPasswordHash(hash string) *UserBuilder {
	b.user.PasswordHash = hash
	return b
}

func (b *UserBuilder) WithRole(role Role) *UserBuilder {
	b.user.Role = role
	return b
}

func (b *UserBuilder) WithReaderID(readerID int) *UserBuilder {
	b.user.ReaderID = &readerID
	return b
}

func (b *UserBuilder) Build() User {
	return b.user
}
