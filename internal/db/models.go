// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type RefreshTokenWhitelist struct {
	ID           int32
	UserID       int32
	RefreshToken pgtype.UUID
	ExpiresAt    pgtype.Timestamp
	CreatedAt    pgtype.Timestamp
}

type SecM struct {
	PrivateKey []byte
}

type User struct {
	ID             int32
	Username       string
	Email          string
	HashedPassword string
	CreatedAt      pgtype.Timestamp
	IsDeleted      pgtype.Bool
	DeletedAt      pgtype.Timestamp
}
