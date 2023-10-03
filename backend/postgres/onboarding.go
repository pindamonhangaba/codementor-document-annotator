package postgres

import (
	"github.com/jmoiron/sqlx"
)

type UserOnboardingForm struct {
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"password"`
	Name     string `db:"name" json:"name"`
	ApplID   string `json:"applID" binding:"required" example:"app1"`
}

// OnboardingCreator is a service to create new users
type OnboardingCreator struct {
	DB *sqlx.DB
}

// Run creates a new user
func (c *OnboardingCreator) Run(f UserOnboardingForm) (*UserAuthentication, error) {
	u, err := NewUser(f.Email, f.Password, f.Name, f.ApplID)
	if err != nil {
		return nil, err
	}

	user, err := createUser(c.DB, u)
	if err != nil {
		return nil, err
	}

	return user, err
}
