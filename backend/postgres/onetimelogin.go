package postgres

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/pindamonhangaba/monoboi/backend/service"

	sq "github.com/pindamonhangaba/monoboi/backend/lib/sqrl"
)

// CreatorOneTimeLogin service to create new OneTimeLogin
type CreatorOneTimeLogin struct {
	DB  *sqlx.DB
	JWT JWTConfig
}

// Run create new OneTimeLogin
func (c *CreatorOneTimeLogin) Run(otl *service.OneTimeLogin) (*service.OneTimeLogin, error) {
	token, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "gen uuid")
	}
	otl.Token = token
	res, err := createOneTimeLogin(c.DB, otl)
	return res, err
}

// DeleterOneTimeLogin service to return OneTimeLogin
type DeleterOneTimeLogin struct {
	DB *sqlx.DB
}

// Run update new OneTimeLogin
func (u *DeleterOneTimeLogin) Run(otloID int64) (*service.OneTimeLogin, error) {
	res, err := deleteOneTimeLogin(u.DB, otloID)
	return res, err
}

// Create a new OneTimeLogin to database
func createOneTimeLogin(db service.DB, otl *service.OneTimeLogin) (*service.OneTimeLogin, error) {
	query := psql.Insert("one_time_login").
		Columns("user_id", "token", "valid_until", "deleted_at").
		Values(otl.UserID, otl.Token, otl.ValidUntil, otl.DeletedAt).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating create One Time Login sql")
	}

	err = db.Get(otl, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating One Time Login")
	}
	return otl, nil
}

// Delete OneTimeLogin to database by otlo_id
func deleteOneTimeLogin(db service.DB, otloID int64) (*service.OneTimeLogin, error) {
	otl := service.OneTimeLogin{}
	query := psql.Update("one_time_login").
		Set("deleted_at", time.Now()).
		Suffix("RETURNING *").
		Where(sq.Eq{"otlo_id": otloID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating One Time Login delete sql")
	}

	err = db.Get(&otl, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error deleting One Time Login")
	}

	return &otl, nil
}

func validOneTimeLogin(db service.DB, token string) (*ValidOneTime, error) {
	res := ValidOneTime{}
	query := psql.Select("valid, user_id").
		From(`valid_one_time`).
		Prefix(
			`WITH check_one_time AS (
				SELECT otlo_id, user_id, token, valid_until, created_at, deleted_at
				FROM one_time_login
				WHERE token = ?
				AND deleted_at IS NULL
				AND valid_until > NOW()
			),
			valid_one_time AS (
				SELECT user_id,
					CASE
						WHEN count(*) IS NULL THEN FALSE
						WHEN count(*) > 0 THEN TRUE
						ELSE FALSE
					END as "valid"
				FROM check_one_time
				GROUP BY user_id
			)`, token,
		)

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "generating one time login sql")
	}

	err = db.Get(&res, qSQL, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, service.NewUnauthorizedError()
		}
		return nil, errors.Wrap(err, "getting one time login")
	}
	return &res, nil
}

func authenticateOneTime(db service.DB, jwtConfig JWTConfig, userID *uuid.UUID) (*service.AuthResponse, error) {
	aRes := service.AuthResponse{}

	usr, err := fromUserID(db, *userID)
	if err != nil {
		return nil, err
	}

	claims := service.Claims{
		UserID: usr.User.UserID.String(),
		ApplID: usr.ApplID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtConfig.HoursTillExpire)),
		},
	}
	aRes.User = usr.User

	tokenJWT := jwt.NewWithClaims(jwtConfig.SigningMethod, claims)
	jwtToken, err := tokenJWT.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return nil, errors.Wrap(err, "Signing token to string")
	}

	aRes.Jwt = jwtToken

	return &aRes, nil
}
