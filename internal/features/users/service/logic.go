package service

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	pUsers "github.com/dwiw96/ran-user-management/internal/features/users"
	middleware "github.com/dwiw96/ran-user-management/pkg/middleware"
	password "github.com/dwiw96/ran-user-management/pkg/utils/password"
	errs "github.com/dwiw96/ran-user-management/pkg/utils/responses"
)

type usersService struct {
	repo  pUsers.IRepository
	cache pUsers.ICache
	ctx   context.Context
}

func NewUsersService(repo pUsers.IRepository, cache pUsers.ICache, ctx context.Context) pUsers.IService {
	return &usersService{
		repo:  repo,
		cache: cache,
		ctx:   ctx,
	}
}

var (
	errRefreshTokenExp = errors.New("refresh token is expires")
)

func handleError(arg error) (code int, err error) {
	if errors.Is(arg, pgx.ErrNoRows) {
		return errs.CodeFailedUnauthorized, errs.ErrNoData
	}
	var pgErr *pgconn.PgError
	if errors.As(arg, &pgErr) {
		if pgErr.ConstraintName == "ck_transactions_balance" {
			return errs.CodeFailedUser, fmt.Errorf("balance minimum is 0")
		}
		switch pgErr.Code {
		case "23505": // UNIQUE violation
			return errs.CodeFailedDuplicated, errs.ErrDuplicate
		case "23514": // CHECK violation
			return errs.CodeFailedUser, errs.ErrCheckConstraint
		case "23502": // NOT NULL violation
			return errs.CodeFailedUser, errs.ErrNotNull
		case "23503": // Foreign Key violation
			return errs.CodeFailedUser, errs.ErrViolation
		default:
			err = fmt.Errorf("database error occurred")
		}
	}

	return errs.CodeFailedServer, err
}

func (s *usersService) SignUp(input pUsers.SignupRequest) (user *pUsers.User, code int, err error) {
	if input.Email == "" {
		return nil, errs.CodeFailedUser, errs.ErrInvalidInput
	}
	// check if the email have registered
	resGetUser, err := s.repo.GetUserByEmail(s.ctx, input.Email)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return nil, errs.CodeFailedServer, err
	}
	if resGetUser.Email == input.Email && !resGetUser.IsDeleted.Bool {
		return nil, errs.CodeFailedDuplicated, fmt.Errorf("this email address is already in use")
	}

	// create arg for create user repository
	arg := pUsers.CreateUserParams{
		Username:       input.Username,
		Email:          input.Email,
		HashedPassword: input.Password,
	}

	arg.HashedPassword, err = password.HashingPassword(input.Password)
	if err != nil {
		return nil, errs.CodeFailedServer, err
	}

	if resGetUser.IsDeleted.Bool {
		user, err = s.repo.UpdateUserIsDeleted(s.ctx, arg)
	} else {
		user, err = s.repo.CreateUser(s.ctx, arg)
	}

	if err != nil {
		code, err = handleError(err)
		return nil, code, err
	}

	return user, errs.CodeSuccessCreate, nil
}

func (s *usersService) LogIn(input pUsers.LoginRequest) (user *pUsers.User, accessToken, refreshToken string, code int, err error) {
	user, err = s.repo.GetUserByEmail(s.ctx, input.Email)
	if err != nil {
		code, err = handleError(err)
		return nil, "", "", code, err
	}

	err = password.VerifyHashPassword(input.Password, user.HashedPassword)
	if err != nil {
		errMsg := errors.New("password is wrong")
		return nil, "", "", errs.CodeFailedUnauthorized, errMsg
	}

	key, err := s.repo.LoadKey(s.ctx)
	if err != nil {
		return nil, "", "", errs.CodeFailedServer, fmt.Errorf("load key error: %w", err)
	}

	accessToken, err = middleware.CreateToken(*user, 60, key)
	if err != nil {
		errMsg := errors.New("failed generate access token")
		return nil, "", "", errs.CodeFailedServer, errMsg
	}
	refreshTokenUUID, err := uuid.NewRandom()
	if err != nil {
		errMsg := errors.New("failed generate refresh token")
		return nil, "", "", errs.CodeFailedServer, errMsg
	}

	insertRefreshTokenArg := pUsers.InsertRefreshTokenParams{
		UserID:       user.ID,
		RefreshToken: pgtype.UUID{Bytes: refreshTokenUUID, Valid: true},
	}
	err = s.repo.InsertRefreshToken(s.ctx, insertRefreshTokenArg)
	if err != nil {
		code, err = handleError(err)
		return nil, "", "", code, err
	}

	refreshToken = refreshTokenUUID.String()

	return user, accessToken, refreshToken, errs.CodeSuccess, nil
}

func (s *usersService) LogOut(payload pUsers.JwtPayload) error {
	err := s.repo.DeleteRefreshToken(s.ctx, payload.UserID)
	if err != nil {
		return err
	}

	err = s.cache.CachingBlockedToken(payload)

	return err
}

func (s *usersService) DeleteUser(arg pUsers.SoftDeleteUserParams) (code int, err error) {
	err = s.repo.DeleteUserTx(s.ctx, arg)
	if err != nil {
		return errs.CodeFailedUser, err
	}

	return errs.CodeSuccess, nil
}

func (s *usersService) RefreshToken(refreshToken, accessToken string) (newRefreshToken, newAccessToken string, code int, err error) {
	key, err := s.repo.LoadKey(s.ctx)
	if err != nil {
		return "", "", errs.CodeFailedServer, err
	}

	authHeader := "Bearer " + accessToken
	payload, err := middleware.ReadToken(authHeader, key)
	if err != nil {
		return "", "", errs.CodeFailedServer, err
	}

	err = s.cache.CachingBlockedToken(*payload)
	if err != nil {
		return "", "", errs.CodeFailedServer, fmt.Errorf("failed to caching access token, msg: %v", err)
	}

	refreshTokenUUID, err := uuid.Parse(refreshToken)
	if err != nil {
		return "", "", errs.CodeFailedServer, fmt.Errorf("failed to convert refresh token from string to uuid, msg: %v", err)
	}

	// Read and validate refresh token from database
	getRefreshTokenArg := pUsers.GetRefreshTokenParams{
		UserID:       payload.UserID,
		RefreshToken: pgtype.UUID{Bytes: refreshTokenUUID, Valid: true},
	}
	resGetRefreshToken, errReadRefreshToken := s.repo.GetRefreshToken(s.ctx, getRefreshTokenArg)
	err = s.validateRefreshToken(resGetRefreshToken, errReadRefreshToken)
	if err != nil {
		return "", "", errs.CodeFailedUnauthorized, err
	}

	// When the refresh token is expired:
	// delete refresh token from database
	if time.Now().UTC().After(resGetRefreshToken.ExpiresAt.Time) {
		err = s.deleteRefreshToken(resGetRefreshToken)
		if err != nil {
			return "", "", errs.CodeFailedServer, err
		}

		return "", "", errs.CodeFailedUnauthorized, errRefreshTokenExp
	}

	var newRefreshTokenUUID uuid.UUID
	newAccessToken, newRefreshTokenUUID, err = s.createNewToken(key, payload)
	if err != nil {
		return "", "", errs.CodeFailedServer, err
	}
	fmt.Println("new refesh token uuid:", newRefreshTokenUUID)

	newRefreshToken = newRefreshTokenUUID.String()
	fmt.Println("new refesh token:", newRefreshToken)

	return newRefreshToken, newAccessToken, errs.CodeSuccess, nil
}

// ValidateRefreshToken return error.
//
// ValidateRefreshToken check the refresh token from database, what to check:
//   - check error when read from database.
//   - check if the refresh token is nil
//   - check if refresh token is expired
func (s *usersService) validateRefreshToken(arg *pUsers.RefreshTokenWhitelist, errIn error) (err error) {
	if errIn != nil {
		return fmt.Errorf("invalid refresh token, msg: %v", errIn)
	}
	if !arg.RefreshToken.Valid {
		return fmt.Errorf("invalid refresh token")
	}

	return nil
}

func (s *usersService) deleteRefreshToken(arg *pUsers.RefreshTokenWhitelist) (err error) {
	if time.Now().UTC().After(arg.ExpiresAt.Time) {
		err = s.repo.DeleteRefreshToken(s.ctx, arg.UserID)
		if err != nil {
			return fmt.Errorf("failed to process expired refresh token, msg: %v", err)
		}
		return fmt.Errorf("refresh token is expire")
	}

	return nil
}

// createNewToken return new access token, new refresh token and error
func (s *usersService) createNewToken(key *rsa.PrivateKey, payload *pUsers.JwtPayload) (newAccessToken string, newRefreshTokenUUID uuid.UUID, err error) {
	user := pUsers.User{
		ID:       payload.UserID,
		Username: payload.Name,
		Email:    payload.Email,
	}

	newAccessToken, err = middleware.CreateToken(user, 60, key)
	if err != nil {
		return
	}
	newRefreshTokenUUID, err = uuid.NewRandom()
	if err != nil {
		return
	}
	err = s.repo.UpdateRefreshToken(s.ctx, payload.UserID, newRefreshTokenUUID)
	if err != nil {
		return
	}

	return
}
