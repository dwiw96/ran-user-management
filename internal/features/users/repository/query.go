package repository

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"

	db "github.com/dwiw96/ran-user-management/internal/db"
	pUsers "github.com/dwiw96/ran-user-management/internal/features/users"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type usersRepository struct {
	db   db.DBTX
	txDb *pgxpool.Pool
}

func NewUsersRepository(db db.DBTX, txDb *pgxpool.Pool) pUsers.IRepository {
	return &usersRepository{
		db:   db,
		txDb: txDb,
	}
}

type transactionTx struct {
	db *pgxpool.Pool
}

func NewTransactionTx(db *pgxpool.Pool) *transactionTx {
	return &transactionTx{
		db: db,
	}
}

func (r *usersRepository) ExecDbTx(ctx context.Context, fn func(*usersRepository) error) error {
	tx, err := r.txDb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start db transaction, err: %v", err)
	}

	q := &usersRepository{db: tx}
	err = fn(q)
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users(
    username,
    email,
    hashed_password
) VALUES (
    $1, $2, $3
) RETURNING id, username, email, hashed_password, created_at, is_deleted, deleted_at
`

func (q *usersRepository) CreateUser(ctx context.Context, arg pUsers.CreateUserParams) (*pUsers.User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Username, arg.Email, arg.HashedPassword)
	var i pUsers.User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.HashedPassword,
		&i.CreatedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return &i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT 
	id, username, email, hashed_password, created_at, is_deleted, deleted_at 
FROM 
	users 
WHERE email = $1
`

func (q *usersRepository) GetUserByEmail(ctx context.Context, email string) (*pUsers.User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i pUsers.User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.HashedPassword,
		&i.CreatedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return &i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE
    users
SET
    username = coalesce($1, username),
    hashed_password = coalesce($2, hashed_password)
WHERE
    id = $3
AND (
    $1::VARCHAR IS NOT NULL AND $1 IS DISTINCT FROM username OR
    $2::VARCHAR IS NOT NULL AND $2 IS DISTINCT FROM hashed_password
) AND 
    is_deleted = FALSE
RETURNING id, username, email, hashed_password, created_at, is_deleted, deleted_at
`

func (q *usersRepository) UpdateUser(ctx context.Context, arg pUsers.UpdateUserParams) (*pUsers.User, error) {
	row := q.db.QueryRow(ctx, updateUser, arg.Username, arg.HashedPassword, arg.ID)
	var i pUsers.User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.HashedPassword,
		&i.CreatedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return &i, err
}

const softDeleteUser = `-- name: SoftDeleteUser :exec
UPDATE
    users
SET
    is_deleted = TRUE,
    deleted_at = NOW()
WHERE
    id = $1
AND email = $2
AND is_deleted = FALSE
`

func (q *usersRepository) SoftDeleteUser(ctx context.Context, arg pUsers.SoftDeleteUserParams) error {
	res, err := q.db.Exec(ctx, softDeleteUser, arg.ID, arg.Email)

	if res.RowsAffected() == 0 {
		return pUsers.ErrNoRowsAffected
	}

	return err
}

func (r *usersRepository) LoadKey(ctx context.Context) (key *rsa.PrivateKey, err error) {
	query := "select private_key from sec_m"
	var keyBytes []byte
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		errMsg := fmt.Errorf("failed to load private key, err: %v", err)
		return nil, errMsg
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&keyBytes)
		if err != nil {
			errMsg := fmt.Errorf("failed to scan private key, err: %v", err)
			return nil, errMsg
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
		if err != nil {
			errMsg := fmt.Errorf("failed to parse private key, err: %v", err)
			return nil, errMsg
		}

		return privateKey, nil
	}

	return nil, errors.New("no private key found in database")
}

const insertRefreshToken = `-- name: InsertRefreshToken :exec
INSERT INTO 
    refresh_token_whitelist(user_id, refresh_token, expires_at) 
VALUES(
    $1, $2, NOW() + INTERVAL '5 minute'
)
`

func (q *usersRepository) InsertRefreshToken(ctx context.Context, arg pUsers.InsertRefreshTokenParams) error {
	res, err := q.db.Exec(ctx, insertRefreshToken, arg.UserID, arg.RefreshToken)

	if res.RowsAffected() == 0 {
		return pUsers.ErrNoRowsAffected
	}

	return err
}

const getRefreshToken = `-- name: GetRefreshToken :one
SELECT id, user_id, refresh_token, expires_at, created_at FROM refresh_token_whitelist WHERE user_id = $1 AND refresh_token = $2
`

func (q *usersRepository) GetRefreshToken(ctx context.Context, arg pUsers.GetRefreshTokenParams) (*pUsers.RefreshTokenWhitelist, error) {
	row := q.db.QueryRow(ctx, getRefreshToken, arg.UserID, arg.RefreshToken)
	var i pUsers.RefreshTokenWhitelist
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshToken,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return &i, err
}

const deleteRefreshToken = `-- name: DeleteRefreshToken :exec
DELETE FROM refresh_token_whitelist WHERE user_id = $1
`

func (q *usersRepository) DeleteRefreshToken(ctx context.Context, userID int32) error {
	res, err := q.db.Exec(ctx, deleteRefreshToken, userID)

	if res.RowsAffected() == 0 {
		return pUsers.ErrNoRowsAffected
	}

	return err
}

func (r *usersRepository) UpdateRefreshToken(ctx context.Context, userID int32, refreshToken uuid.UUID) (err error) {
	r.ExecDbTx(ctx, func(tr *usersRepository) error {
		err = tr.DeleteRefreshToken(ctx, userID)
		if err != nil {
			return err
		}

		arg := pUsers.InsertRefreshTokenParams{
			UserID:       userID,
			RefreshToken: pgtype.UUID{Bytes: refreshToken, Valid: true},
		}
		err = tr.InsertRefreshToken(ctx, arg)
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}

const updateUserIsDeleted = `-- name: UpdateUserIsDeleted :one
UPDATE
    users
SET
    is_deleted = FALSE,
	username = $1,
	hashed_password = $2,
	created_at = NOW()
WHERE
    email = $3
AND
	is_deleted = TRUE
RETURNING id, username, email, hashed_password, created_at, is_deleted, deleted_at
`

func (q *usersRepository) UpdateUserIsDeleted(ctx context.Context, arg pUsers.CreateUserParams) (*pUsers.User, error) {
	row := q.db.QueryRow(ctx, updateUserIsDeleted, arg.Username, arg.HashedPassword, arg.Email)
	var i pUsers.User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.HashedPassword,
		&i.CreatedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return &i, err
}

func (r *usersRepository) DeleteUserTx(ctx context.Context, arg pUsers.SoftDeleteUserParams) (err error) {
	err = r.ExecDbTx(ctx, func(ar *usersRepository) error {
		err = ar.DeleteRefreshToken(ctx, arg.ID)
		if err != nil && err != pUsers.ErrNoRowsAffected {
			return err
		}

		err = ar.SoftDeleteUser(ctx, arg)
		if err != nil {
			return err
		}

		return nil
	})
	return err
}
