package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/authorizer"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
			user_id integer NOT NULL REFERENCES users(user_id), 
			order_number text, 
			order_time timestamptz (0) NOT NULL DEFAULT CURRENT_TIMESTAMP, 
			order_status text,
			points_accrued real NOT NULL DEFAULT 0,
			PRIMARY KEY(order_number)
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS points_used (
			user_id integer NOT NULL REFERENCES users(user_id), 
			order_number text, 
			points real NOT NULL,
			time_of_used timestamptz (0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(order_number)
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS accounts (
			user_id integer REFERENCES users(user_id), 
			balance real DEFAULT 0,
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

	result, err = db.dbHandle.ExecContext(ctx,
		`INSERT INTO accounts (user_id) 
		VALUES ((SELECT user_id FROM users WHERE login = $1))`,
		regData.Login)
	if err != nil {
		return err
	}
	rows, err = result.RowsAffected()
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

func (db *DBStorage) PostUserOrder(ctx context.Context, orderNumber string, userLogin string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := db.dbHandle.ExecContext(ctx,
		`INSERT INTO orders (user_id , order_number, order_status) 
		VALUES ((SELECT user_id FROM users WHERE login = $1), $2, $3)`,
		userLogin, orderNumber, NewOrder)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			row := db.dbHandle.QueryRowContext(ctx,
				`SELECT login FROM users INNER JOIN orders 
				ON users.user_id = orders.user_id 
				WHERE orders.order_number = $1`, orderNumber)
			var userOrderLogin string
			errScan := row.Scan(&userOrderLogin)
			if errScan != nil {
				return err
			}
			if userOrderLogin != userLogin {
				return NewStorError(UploadByAnotherUser, err)
			}
			return NewStorError(UploadByThisUser, err)
		}
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

func (db *DBStorage) GetUserOrders(ctx context.Context, userLogin string) ([]ResponseOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := db.dbHandle.QueryContext(ctx,
		`SELECT o.order_number, o.order_time, o.order_status, o.points_accrued 
		FROM orders o INNER JOIN users u
		ON o.user_id = u.user_id
		WHERE u.login = $1
		ORDER BY o.order_time`, userLogin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var respOrders []ResponseOrder
	for rows.Next() {
		var ord ResponseOrder
		err = rows.Scan(&ord.Number, &ord.UploadedTime, &ord.Status, &ord.Accrual)
		if err != nil {
			return nil, err
		}
		respOrders = append(respOrders, ord)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return respOrders, nil
}

func (db *DBStorage) GetUserBalance(ctx context.Context, userLogin string) (ResponseBalance, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	row := db.dbHandle.QueryRowContext(ctx,
		`SELECT a.balance, coalesce(sum(pu.points), 0)  
		FROM accounts a LEFT JOIN points_used pu 
		ON a.user_id = pu.user_id 
		WHERE a.user_id = (SELECT user_id FROM users WHERE login = $1) 
		GROUP BY a.balance`, userLogin)
	var respBalance ResponseBalance
	err := row.Scan(&respBalance.PointsBalance, &respBalance.PointsUsed)
	if err != nil {
		return ResponseBalance{}, err
	}

	return respBalance, nil
}

func (db *DBStorage) PostWithdraw(ctx context.Context, userLogin string, withdrawData RequestWithdrawData) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := db.dbHandle.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRowContext(ctx,
		`SELECT a.balance 
		FROM accounts a INNER JOIN users u 
		ON a.user_id = u.user_id 
		WHERE u.login = $1`, userLogin)
	var balance float32
	err = row.Scan(&balance)
	if err != nil {
		tx.Commit()
		return err
	}
	if balance < withdrawData.WithdrawSum {
		tx.Commit()
		return NewStorError(NotEnoughPoints, nil)
	}

	result, err := tx.ExecContext(ctx,
		`INSERT INTO points_used (user_id, order_number, points) 
		VALUES ((SELECT user_id FROM users WHERE login = $1), $2, $3)`,
		userLogin, withdrawData.OrderNumber, withdrawData.WithdrawSum)
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows != 1 {
		tx.Rollback()
		return errors.New("expected to affect 1 row")
	}

	result, err = tx.ExecContext(ctx,
		`UPDATE accounts SET balance = $1 
		WHERE user_id = (SELECT user_id FROM users WHERE login = $2)`,
		balance-withdrawData.WithdrawSum, userLogin)
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err = result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows != 1 {
		tx.Rollback()
		return errors.New("expected to affect 1 row")
	}

	return tx.Commit()
}

func (db *DBStorage) GetUserWithdrawals(ctx context.Context, userLogin string) ([]ResponseWithdrawals, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := db.dbHandle.QueryContext(ctx,
		`SELECT pu.order_number, pu.points, pu.time_of_used 
		FROM points_used pu INNER JOIN users u ON pu.user_id = u.user_id 
		WHERE u.login = $1 
		ORDER BY pu.time_of_used`, userLogin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var respWithdraw []ResponseWithdrawals
	for rows.Next() {
		var withdraw ResponseWithdrawals
		err = rows.Scan(&withdraw.OrderNumber, &withdraw.WithdrawSum, &withdraw.WithdrawTime)
		if err != nil {
			return nil, err
		}
		respWithdraw = append(respWithdraw, withdraw)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return respWithdraw, nil
}

func (db *DBStorage) UpdateUserAccrual(ctx context.Context, newData ResponseAccrual) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := db.dbHandle.Begin()
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx,
		`UPDATE orders 
		SET order_status = $1, points_accrued = $2 
		WHERE order_number = $3`,
		newData.OrderStatus, newData.Accrual, newData.OrderNumber)
	if err != nil {
		tx.Rollback()
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows != 1 {
		tx.Rollback()
		return errors.New("expected to affect 1 row")
	}

	result, err = tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance + $1 
		WHERE user_id = (SELECT user_id FROM orders WHERE order_number = $2)`,
		newData.Accrual, newData.OrderNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err = result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows != 1 {
		tx.Rollback()
		return errors.New("expected to affect 1 row")
	}

	return tx.Commit()
}
