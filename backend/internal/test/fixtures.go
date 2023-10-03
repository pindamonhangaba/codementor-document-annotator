package test

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/gofrs/uuid"
	"github.com/pindamonhangaba/monoboi/backend/service"
)

// FixtureMap represents the main definition which fixtures are available though Fixtures()
type FixtureMap struct {
	User1                *service.User
	User2                *service.User
	AuthenticationEmail1 *service.AuthenticationEmail
}

// Fixtures returns a function wrapping our fixtures, which tests are allowed to manipulate.
// Each test (which may run concurrently) receives a fresh copy, preventing side effects between test runs.
func Fixtures() FixtureMap {
	createdAt, err := time.Parse("2006-01-02", "2019-01-31")
	if err != nil {
		panic("time should parse correctly")
	}

	user1 := &service.User{
		UserID:     uuid.Must(uuid.FromString("45cac6c9-bd9a-4f87-89e2-048bb1d855e6")),
		Name:       " user",
		PushTokens: service.PushTokens{},
		Info:       service.UserInfo{},
		ApplID:     "api",
		CreatedAt:  createdAt,
	}

	user2 := &service.User{
		UserID:     uuid.Must(uuid.FromString("d975cc1e-40c4-4524-bbc6-d6aadf9a6e54")),
		Name:       "common user",
		PushTokens: service.PushTokens{},
		Info:       service.UserInfo{},
		ApplID:     "api",
		CreatedAt:  createdAt,
	}

	authenticationEmail1 := &service.AuthenticationEmail{
		AuemID:    0,
		UserID:    user1.UserID,
		Email:     "user@example.com.br",
		Password:  []byte("1qw3e4r"),
		CreatedAt: createdAt,
	}

	return FixtureMap{
		User1:                user1,
		User2:                user2,
		AuthenticationEmail1: authenticationEmail1,
	}
}

// Inserts defines the order in which the fixtures will be inserted
// into the test database
func Inserts() []service.SQLer {
	f := Fixtures()

	return []service.SQLer{
		f.User1.AsFixture(),
		f.AuthenticationEmail1.AsFixture(),
	}
}

type Fixture struct {
	DB *sql.DB
}

func (f Fixture) Run() error {
	ins := Inserts()

	tx, err := f.DB.Begin()
	if err != nil {
		return err
	}

	err = func() error {
		for _, sqler := range ins {
			s, args := sqler.SQL()
			res, err := tx.Exec(s, args...)
			if err != nil {
				return err
			}
			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return errors.New("insert fixture " + s)
			}
			s, args = sqler.Cleanup()
			if len(s) == 0 {
				continue
			}
			_, err = tx.Exec(s, args...)
			if err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

func StrPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func Int8Ptr(i int8) *int8 {
	return &i
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func FloatPtr(f float64) *float64 {
	return &f
}

func BoolPtr(f bool) *bool {
	return &f
}
