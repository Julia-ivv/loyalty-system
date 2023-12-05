package storage

import (
	"context"
	"database/sql"
	"time"
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
			password text NOT NULL, 
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
			points_accrued integer NOT NULL,
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
