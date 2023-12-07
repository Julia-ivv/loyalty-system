package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/authorizer"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	dbHandle *sql.DB
}

func NewDBStorage(DBURI string) (*DBStorage, error) {
	db, err := sql.Open("pgx", DBURI)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS users (
			user_id serial, 
			login text UNIQUE NOT NULL, 
			hash text NOT NULL,
			salt text NOT NULL, 
			PRIMARY KEY(user_id)
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS orders (
			user_id integer REFERENCES users(user_id), 
			order_number integer, 
			order_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, 
			order_status text NOT NULL,
			points_accrued integer NOT NULL DEFAULT 0,
			PRIMARY KEY(order_number)
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS points_used (
			user_id integer REFERENCES users(user_id), 
			order_number integer, 
			points integer NOT NULL,
			time_of_used timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(order_number)
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS accounts (
			user_id integer REFERENCES users(user_id), 
			balance integer,
			PRIMARY KEY(user_id)
		)`)
	if err != nil {
		return nil, err
	}

	return &DBStorage{dbHandle: db}, nil
}

func (db *DBStorage) Close() error {
	return db.dbHandle.Close()
}

func (db *DBStorage) RegUser(ctx context.Context, regData RequestRegData) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	salt, err := GenerateRandomString(LengthSalt)
	if err != nil {
		return err
	}
	result, err := db.dbHandle.ExecContext(ctx,
		"INSERT INTO users (login, hash, salt) VALUES ($1, $2, $3)",
		regData.Login, hash(regData.Pwd, salt), salt)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("expected to affect 1 row")
	}

	return nil
}

func (db *DBStorage) AuthUser(ctx context.Context, authData RequestAuthData) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	row := db.dbHandle.QueryRowContext(ctx,
		"SELECT hash, salt FROM users WHERE login=$1", authData.Login)

	var dbHash, dbSalt string
	err := row.Scan(&dbHash, &dbSalt)
	if err != nil {
		return authorizer.NewAuthError(authorizer.QeuryError, err)
	}

	newHash := hash(authData.Pwd, dbSalt)
	if newHash != dbHash {
		return authorizer.NewAuthError(authorizer.InvalidHash, errors.New("invalid hash"))
	}

	return nil
}
