package postgres

import (
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pindamonhangaba/monoboi/backend/internal/test"
	"golang.org/x/crypto/bcrypt"
)

var usr = UserOnboardingForm{
	Name:     "Test User",
	Email:    "onboardingpatient@notproduction.space",
	Password: "1q2w3e4r",
}

func TestAccountCreate(t *testing.T) {
	test.WithTestDatabase(t, func(db *sql.DB) {
		dbx := sqlx.NewDb(db, "postgres")
		onb := OnboardingCreator{DB: dbx}

		r, err := onb.Run(usr)
		if err != nil {
			t.Error(err)
		}

		err = bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(r.Password))
		if err != nil {
			t.Error(err)
		}
	})
}

func TestFailedOnboarding(t *testing.T) {
	test.WithTestDatabase(t, func(db *sql.DB) {
		dbx := sqlx.NewDb(db, "postgres")
		onb := OnboardingCreator{DB: dbx}

		_, err := onb.Run(UserOnboardingForm{
			Name:     "Test User",
			Email:    "",
			Password: "1q2w3e4r",
		})
		if err == nil {
			t.Errorf("expected email validation error")
		}
	})
}
