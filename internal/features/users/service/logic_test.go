package service

import (
	"context"
	"os"

	"strings"
	"testing"

	pUsers "github.com/dwiw96/ran-user-management/internal/features/users"
	cache "github.com/dwiw96/ran-user-management/internal/features/users/cache"
	repo "github.com/dwiw96/ran-user-management/internal/features/users/repository"

	middleware "github.com/dwiw96/ran-user-management/pkg/middleware"
	generator "github.com/dwiw96/ran-user-management/pkg/utils/generator"
	password "github.com/dwiw96/ran-user-management/pkg/utils/password"
	errs "github.com/dwiw96/ran-user-management/pkg/utils/responses"
	testUtils "github.com/dwiw96/ran-user-management/testutils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	serviceTest pUsers.IService
	poolTest    *pgxpool.Pool
	ctx         context.Context
	repoTest    pUsers.IRepository
)

func TestMain(m *testing.M) {
	poolTest = testUtils.GetPool()
	ctx = testUtils.GetContext()

	schemaCleanup := testUtils.SetupDB("test_service_auth")

	password.JwtInit(poolTest, ctx)

	client := testUtils.GetRedisClient()

	repoTest = repo.NewUsersRepository(poolTest, poolTest)
	cacheTest := cache.NewUsersCache(client, ctx)
	serviceTest = NewUsersService(repoTest, cacheTest, ctx)

	exitTest := m.Run()

	schemaCleanup()
	poolTest.Close()
	ctx.Done()
	client.Close()

	os.Exit(exitTest)
}

func createUser(t *testing.T) (user *pUsers.User, signupReq pUsers.SignupRequest) {
	email := generator.CreateRandomEmail(generator.CreateRandomString(5))

	arg := pUsers.SignupRequest{
		Username: generator.CreateRandomString(5),
		Email:    email,
		Password: generator.CreateRandomString(10),
	}
	assert.NotEmpty(t, arg.Username)
	assert.NotEmpty(t, arg.Email)
	assert.NotEmpty(t, arg.Password)

	res, code, err := serviceTest.SignUp(arg)

	require.NoError(t, err)
	require.Equal(t, errs.CodeSuccessCreate, code)
	assert.NotZero(t, res.ID)
	assert.Equal(t, arg.Username, res.Username)
	assert.Equal(t, arg.Email, res.Email)
	assert.NotEqual(t, arg.Password, res.HashedPassword)
	assert.False(t, res.CreatedAt.Time.IsZero())
	assert.False(t, res.IsDeleted.Bool)
	assert.False(t, res.DeletedAt.Valid)

	return res, arg
}

func TestSignUp(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	email := generator.CreateRandomEmail(generator.CreateRandomString(5))
	testCases := []struct {
		desc string
		arg  pUsers.SignupRequest
		user *pUsers.User
		code int
		err  bool
	}{
		{
			desc: "success_new",
			arg: pUsers.SignupRequest{
				Username: generator.CreateRandomString(5),
				Email:    email,
				Password: generator.CreateRandomString(7),
			},
			code: errs.CodeSuccessCreate,
			err:  false,
		}, {
			desc: "failed_empty_username",
			arg: pUsers.SignupRequest{
				Email:    generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Password: generator.CreateRandomString(7),
			},
			code: errs.CodeFailedUser,
			err:  true,
		}, {
			desc: "failed_empty_password",
			arg: pUsers.SignupRequest{
				Username: generator.CreateRandomString(5),
				Email:    generator.CreateRandomEmail(generator.CreateRandomString(5)),
			},
			code: errs.CodeFailedServer,
			err:  true,
		}, {
			desc: "failed_empty_email",
			arg: pUsers.SignupRequest{
				Username: generator.CreateRandomString(5),
				Password: generator.CreateRandomString(7),
			},
			code: errs.CodeFailedUser,
			err:  true,
		}, {
			desc: "failed_duplicate_email",
			arg: pUsers.SignupRequest{
				Username: generator.CreateRandomString(5),
				Email:    email,
				Password: generator.CreateRandomString(10),
			},
			code: errs.CodeFailedDuplicated,
			err:  true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.SignUp(tC.arg)
			require.Equal(t, tC.code, code)

			if !tC.err {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, tC.arg.Username, res.Username)
				assert.Equal(t, tC.arg.Email, res.Email)
				assert.NotEqual(t, tC.arg.Password, res.HashedPassword)
				assert.False(t, res.CreatedAt.Time.IsZero())
				assert.False(t, res.IsDeleted.Bool)
				assert.False(t, res.DeletedAt.Valid)
			} else {
				require.Error(t, err)
			}
		})
	}

	var users []*pUsers.User
	for i := 0; i < 2; i++ {
		username := generator.CreateRandomString(5)
		createUserArg := pUsers.CreateUserParams{
			Username:       username,
			Email:          generator.CreateRandomEmail(username),
			HashedPassword: generator.CreateRandomString(10),
		}
		user, err := repoTest.CreateUser(ctx, createUserArg)
		require.NoError(t, err)
		assert.NotNil(t, user)

		users = append(users, user)
	}

	deleteUserArg := pUsers.SoftDeleteUserParams{
		ID:    users[0].ID,
		Email: users[0].Email,
	}
	err = repoTest.SoftDeleteUser(ctx, deleteUserArg)
	require.NoError(t, err)

	testCases = []struct {
		desc string
		arg  pUsers.SignupRequest
		user *pUsers.User
		code int
		err  bool
	}{
		{
			desc: "success_reregister",
			arg: pUsers.SignupRequest{
				Username: users[0].Username,
				Email:    users[0].Email,
				Password: users[0].HashedPassword,
			},
			user: users[0],
			code: errs.CodeSuccessCreate,
			err:  false,
		}, {
			desc: "failed_is_delete",
			arg: pUsers.SignupRequest{
				Username: users[1].Username,
				Email:    users[1].Email,
				Password: users[1].HashedPassword,
			},
			user: users[1],
			code: errs.CodeFailedDuplicated,
			err:  true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.SignUp(tC.arg)
			require.Equal(t, tC.code, code)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, tC.user.ID, res.ID)
				assert.Equal(t, tC.arg.Username, res.Username)
				assert.Equal(t, tC.arg.Email, res.Email)
				assert.NotEqual(t, tC.arg.Password, res.HashedPassword)
				assert.False(t, res.CreatedAt.Time.IsZero())
				assert.False(t, res.IsDeleted.Bool)
				assert.False(t, res.DeletedAt.Time.IsZero())
				assert.True(t, res.DeletedAt.Valid)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestLogIn(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	user, signUpReq := createUser(t)

	testCases := []struct {
		name string
		arg  pUsers.LoginRequest
		code int
		err  bool
	}{
		{
			name: "success",
			arg: pUsers.LoginRequest{
				Email:    signUpReq.Email,
				Password: signUpReq.Password,
			},
			code: errs.CodeSuccess,
			err:  false,
		}, {
			name: "failed_email_wrong",
			arg: pUsers.LoginRequest{
				Email:    "err" + signUpReq.Email,
				Password: signUpReq.Password,
			},
			code: errs.CodeFailedUnauthorized,
			err:  true,
		}, {
			name: "failed_password_wrong",
			arg: pUsers.LoginRequest{
				Email:    signUpReq.Email,
				Password: "err" + signUpReq.Password,
			},
			code: errs.CodeFailedUnauthorized,
			err:  true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			res, accessToken, refreshToken, code, err := serviceTest.LogIn(tC.arg)
			assert.Equal(t, tC.code, code)
			if !tC.err {
				require.NoError(t, err)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
				user.Username = res.Username

				assert.Equal(t, signUpReq.Username, res.Username)
				assert.Equal(t, tC.arg.Email, res.Email)
				assert.Equal(t, user.HashedPassword, res.HashedPassword)
				assert.False(t, res.CreatedAt.Time.IsZero())
				assert.False(t, res.IsDeleted.Bool)
				assert.False(t, res.DeletedAt.Valid)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestLogOut(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	_, signUpReq := createUser(t)

	argLogin := pUsers.LoginRequest{
		Email:    signUpReq.Email,
		Password: signUpReq.Password,
	}

	_, accessToken, refreshToken, code, err := serviceTest.LogIn(argLogin)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, errs.CodeSuccess, code)

	key, err := repoTest.LoadKey(ctx)
	require.NoError(t, err)
	require.NotNil(t, key)

	payload, err := middleware.ReadToken(accessToken, key)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err = serviceTest.LogOut(*payload)
		require.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		err = serviceTest.LogOut(*payload)
		require.Error(t, err)
	})
}

func insertRefreshTokenTest(t *testing.T, userID int32) uuid.UUID {
	refreshToken, err := uuid.NewRandom()
	require.NoError(t, err)

	arg := pUsers.InsertRefreshTokenParams{
		UserID:       userID,
		RefreshToken: pgtype.UUID{Bytes: refreshToken, Valid: true},
	}
	err = repoTest.InsertRefreshToken(ctx, arg)
	require.NoError(t, err)

	return refreshToken
}

func TestDeleteUser(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	var users []pUsers.User
	for i := 0; i < 5; i++ {
		user, _ := createUser(t)
		insertRefreshTokenTest(t, user.ID)
		users = append(users, *user)
	}

	user6, _ := createUser(t)
	users = append(users, *user6)

	testCases := []struct {
		desc string
		arg  pUsers.SoftDeleteUserParams
		code int
		err  bool
	}{
		{
			desc: "success",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[0].ID,
				Email: users[0].Email,
			},
			code: errs.CodeSuccess,
			err:  false,
		}, {
			desc: "success",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[1].ID,
				Email: users[1].Email,
			},
			code: errs.CodeSuccess,
			err:  false,
		}, {
			desc: "success",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[2].ID,
				Email: users[2].Email,
			},
			code: errs.CodeSuccess,
			err:  false,
		}, {
			desc: "success_without_refreshtoken",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[5].ID,
				Email: users[5].Email,
			},
			code: errs.CodeSuccess,
			err:  false,
		}, {
			desc: "failed_wrong_id",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[3].ID + 5,
				Email: users[3].Email,
			},
			code: errs.CodeFailedUser,
			err:  true,
		}, {
			desc: "failed_wrong_email",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[4].ID,
				Email: "a" + users[4].Email,
			},

			code: errs.CodeFailedUser,
			err:  true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			code, err := serviceTest.DeleteUser(tC.arg)
			assert.Equal(t, tC.code, code)
			if !tC.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	_, signUpReq := createUser(t)

	argLogin := pUsers.LoginRequest{
		Email:    signUpReq.Email,
		Password: signUpReq.Password,
	}

	_, accessToken, refreshToken, code, err := serviceTest.LogIn(argLogin)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, 200, code)
	accessTokenNoBearer := strings.Split(accessToken, " ")

	testCases := []struct {
		desc         string
		refreshToken string
		accessToken  string
		err          bool
	}{
		{
			desc:         "success",
			refreshToken: refreshToken,
			accessToken:  accessTokenNoBearer[1],
			err:          false,
		}, {
			desc:         "failed_invalid_access_token",
			refreshToken: refreshToken,
			accessToken:  accessTokenNoBearer[1] + "a",
			err:          true,
		}, {
			desc:         "failed_invalid_refresh_token",
			refreshToken: refreshToken + "a",
			accessToken:  accessTokenNoBearer[1],
			err:          true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			newRefreshToken, newAccessToken, code, err := serviceTest.RefreshToken(tC.refreshToken, tC.accessToken)
			if !tC.err {
				require.NoError(t, err)
				require.Equal(t, 200, code)
				assert.NotEmpty(t, newRefreshToken)
				assert.NotEmpty(t, newAccessToken)
			} else {
				require.Error(t, err)
			}
		})
	}
}
