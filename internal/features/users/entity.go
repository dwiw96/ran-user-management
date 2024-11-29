package users

import (
	"context"
	"crypto/rsa"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNoRowsAffected = errors.New("no rows affected")
)

// database model for users table
type User struct {
	ID             int32
	Username       string
	Email          string
	HashedPassword string
	CreatedAt      pgtype.Timestamp
	IsDeleted      pgtype.Bool
	DeletedAt      pgtype.Timestamp
}

// params for repository method
type CreateUserParams struct {
	Username       string
	Email          string
	HashedPassword string
}

type UpdateUserParams struct {
	Username       string
	HashedPassword string
	ID             int32
}

type SoftDeleteUserParams struct {
	ID    int32
	Email string
}

// params for service method
type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// param for jwt
type JwtPayload struct {
	jwt.RegisteredClaims
	ID      uuid.UUID `json:"id"`
	UserID  int32     `json:"user_id"`
	Iss     string    `json:"iss"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Address string    `json:"address,omitempty"`
	Iat     int64     `json:"iat"`
	Exp     int64     `json:"exp"`
}

// param for refresh token
type RefreshTokenWhitelist struct {
	ID           int32
	UserID       int32
	RefreshToken pgtype.UUID
	ExpiresAt    pgtype.Timestamp
	CreatedAt    pgtype.Timestamp
}

type InsertRefreshTokenParams struct {
	UserID       int32
	RefreshToken pgtype.UUID
}

type GetRefreshTokenParams struct {
	UserID       int32
	RefreshToken pgtype.UUID
}

type IRepository interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error)
	SoftDeleteUser(ctx context.Context, arg SoftDeleteUserParams) error

	LoadKey(ctx context.Context) (*rsa.PrivateKey, error)
	InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) error
	GetRefreshToken(ctx context.Context, arg GetRefreshTokenParams) (*RefreshTokenWhitelist, error)
	DeleteRefreshToken(ctx context.Context, userID int32) (err error)
	UpdateRefreshToken(ctx context.Context, userID int32, refreshToken uuid.UUID) error

	UpdateUserIsDeleted(ctx context.Context, arg CreateUserParams) (*User, error)
	DeleteUserTx(ctx context.Context, arg SoftDeleteUserParams) (err error)
}

type IService interface {
	SignUp(input SignupRequest) (user *User, code int, err error)
	LogIn(input LoginRequest) (user *User, accessToken, refreshToken string, code int, err error)
	LogOut(payload JwtPayload) error
	DeleteUser(arg SoftDeleteUserParams) (code int, err error)
	RefreshToken(refreshToken, accessToken string) (newRefreshToken, newAccessToken string, code int, err error)
}

type ICache interface {
	CachingBlockedToken(payload JwtPayload) error
	CheckBlockedToken(payload JwtPayload) error
}
