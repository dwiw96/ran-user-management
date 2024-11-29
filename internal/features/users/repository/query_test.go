package repository

import (
	"context"
	"os"
	"time"

	"testing"

	pUsers "github.com/dwiw96/ran-user-management/internal/features/users"
	generator "github.com/dwiw96/ran-user-management/pkg/utils/generator"
	password "github.com/dwiw96/ran-user-management/pkg/utils/password"
	testUtils "github.com/dwiw96/ran-user-management/testutils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	repoTest pUsers.IRepository
	poolTest *pgxpool.Pool
	ctx      context.Context
)

func TestMain(m *testing.M) {
	poolTest = testUtils.GetPool()
	ctx = testUtils.GetContext()

	schemaCleanup := testUtils.SetupDB("test_repo_users")

	password.JwtInit(poolTest, ctx)

	repoTest = NewUsersRepository(poolTest, poolTest)

	exitTest := m.Run()

	schemaCleanup()
	poolTest.Close()
	ctx.Done()

	os.Exit(exitTest)
}

func createRandomUser(t *testing.T) (res *pUsers.User) {
	username := generator.CreateRandomString(generator.RandomInt(3, 13))
	arg := pUsers.CreateUserParams{
		Username:       username,
		Email:          generator.CreateRandomEmail(username),
		HashedPassword: generator.CreateRandomString(generator.RandomInt(20, 20)),
	}

	assert.NotEmpty(t, arg.Username)
	assert.NotEmpty(t, arg.Email)
	assert.NotEmpty(t, arg.HashedPassword)

	res, err := repoTest.CreateUser(ctx, arg)
	require.NoError(t, err)
	assert.NotZero(t, res.ID)
	assert.Equal(t, username, res.Username)
	assert.Equal(t, arg.Email, res.Email)
	assert.Equal(t, arg.HashedPassword, res.HashedPassword)
	assert.False(t, res.CreatedAt.Time.IsZero())
	assert.False(t, res.IsDeleted.Bool)
	assert.False(t, res.DeletedAt.Valid)

	return res
}

func TestCreateUser(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	email := generator.CreateRandomEmail(generator.CreateRandomString(5))
	testCases := []struct {
		desc string
		arg  pUsers.CreateUserParams
		err  bool
	}{
		{
			desc: "success",
			arg: pUsers.CreateUserParams{
				Username:       generator.CreateRandomString(5),
				Email:          email,
				HashedPassword: generator.CreateRandomString(60),
			},
			err: false,
		}, {
			desc: "failed_empty_username",
			arg: pUsers.CreateUserParams{
				Email:          generator.CreateRandomEmail(generator.CreateRandomString(5)),
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		}, {
			desc: "failed_empty_email",
			arg: pUsers.CreateUserParams{
				Username:       generator.CreateRandomEmail(generator.CreateRandomString(5)),
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		}, {
			desc: "failed_empty_hashed_password",
			arg: pUsers.CreateUserParams{
				Username: generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Email:    generator.CreateRandomEmail(generator.CreateRandomString(5)),
			},
			err: true,
		}, {
			desc: "failed_duplicate_email",
			arg: pUsers.CreateUserParams{
				Username:       generator.CreateRandomString(5),
				Email:          email,
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.CreateUser(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, tC.arg.Username, res.Username)
				assert.Equal(t, tC.arg.Email, res.Email)
				assert.Equal(t, tC.arg.HashedPassword, res.HashedPassword)
				assert.False(t, res.IsDeleted.Bool)
				assert.False(t, res.DeletedAt.Valid)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	user := createRandomUser(t)

	testCases := []struct {
		desc  string
		email string
		err   bool
	}{
		{
			desc:  "success",
			email: user.Email,
			err:   false,
		},
		{
			desc:  "failed_empty_email",
			email: "",
			err:   true,
		}, {
			desc:  "failed_invalid_email",
			email: "av088@mail.com",
			err:   true,
		}, {
			desc:  "failed_typo_email",
			email: "a" + user.Email,
			err:   true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.GetUserByEmail(ctx, tC.email)
			if !tC.err {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, user.Email, res.Email)
				assert.Equal(t, user.Username, res.Username)
				assert.Equal(t, user.HashedPassword, res.HashedPassword)
				assert.False(t, res.IsDeleted.Bool)
				assert.False(t, res.DeletedAt.Valid)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	user := createRandomUser(t)

	usernameSuccess := generator.CreateRandomString(7)
	hashedPasswordSuccess := generator.CreateRandomString(60)
	hashedPasswordEmptyUsername := generator.CreateRandomString(60)
	usernameEmptyHashedPassword := generator.CreateRandomString(60)
	testCases := []struct {
		desc string
		arg  pUsers.UpdateUserParams
		ans  pUsers.User
		err  bool
	}{
		{
			desc: "success",
			arg: pUsers.UpdateUserParams{
				ID:             user.ID,
				Username:       usernameSuccess,
				HashedPassword: hashedPasswordSuccess,
			},
			ans: pUsers.User{
				ID:             user.ID,
				Username:       usernameSuccess,
				Email:          user.Email,
				HashedPassword: hashedPasswordSuccess,
				CreatedAt:      user.CreatedAt,
				IsDeleted:      user.IsDeleted,
				DeletedAt:      user.DeletedAt,
			},
			err: false,
		}, {
			desc: "failed_empty_username",
			arg: pUsers.UpdateUserParams{
				ID:             user.ID,
				HashedPassword: hashedPasswordEmptyUsername,
			},
			ans: pUsers.User{
				ID:             0,
				Username:       "",
				HashedPassword: hashedPasswordEmptyUsername,
			},
			err: true,
		}, {
			desc: "failed_empty_hashed_password",
			arg: pUsers.UpdateUserParams{
				ID:       user.ID,
				Username: usernameEmptyHashedPassword,
			},
			ans: pUsers.User{
				ID:             0,
				Username:       usernameEmptyHashedPassword,
				HashedPassword: "",
			},
			err: true,
		}, {
			desc: "failed_empty_arg",
			arg: pUsers.UpdateUserParams{
				ID: user.ID,
			},
			ans: pUsers.User{
				ID:             0,
				Username:       "",
				HashedPassword: "",
			},
			err: true,
		}, {
			desc: "failed_wrong_id",
			arg: pUsers.UpdateUserParams{
				ID:             0,
				Username:       generator.CreateRandomString(7),
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.UpdateUser(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, &tC.ans, res)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestSoftDeleteUser(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	var users []pUsers.User
	for i := 0; i < 5; i++ {
		user := createRandomUser(t)
		users = append(users, *user)
	}

	testCases := []struct {
		desc string
		arg  pUsers.SoftDeleteUserParams
		err  bool
	}{
		{
			desc: "success",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[0].ID,
				Email: users[0].Email,
			},
			err: false,
		}, {
			desc: "success",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[1].ID,
				Email: users[1].Email,
			},
			err: false,
		}, {
			desc: "success",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[2].ID,
				Email: users[2].Email,
			},
			err: false,
		}, {
			desc: "failed_wrong_id",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[3].ID + 5,
				Email: users[3].Email,
			},
			err: true,
		}, {
			desc: "failed_wrong_email",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[4].ID,
				Email: "a" + users[4].Email,
			},
			err: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := repoTest.SoftDeleteUser(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestLoadKey(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	res, err := repoTest.LoadKey(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func inserRefreshTokenTest(t *testing.T, userID int32, refreshToken uuid.UUID) {
	arg := pUsers.InsertRefreshTokenParams{
		UserID:       userID,
		RefreshToken: pgtype.UUID{Bytes: refreshToken, Valid: true},
	}

	err := repoTest.InsertRefreshToken(ctx, arg)
	require.NoError(t, err)
}

func TestInsertRefreshToken(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	refreshToken := uuid.New()

	testCases := []struct {
		name string
		arg  pUsers.InsertRefreshTokenParams
		err  bool
	}{
		{
			name: "succes_1",
			arg: pUsers.InsertRefreshTokenParams{
				UserID:       createRandomUser(t).ID,
				RefreshToken: pgtype.UUID{Bytes: refreshToken, Valid: true},
			},
			err: false,
		}, {
			name: "succes_2",
			arg: pUsers.InsertRefreshTokenParams{
				UserID:       createRandomUser(t).ID,
				RefreshToken: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			},
			err: false,
		}, {
			name: "succes_3",
			arg: pUsers.InsertRefreshTokenParams{
				UserID:       createRandomUser(t).ID,
				RefreshToken: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			},
			err: false,
		}, {
			name: "error_wrong_id",
			arg: pUsers.InsertRefreshTokenParams{
				UserID:       createRandomUser(t).ID + 5,
				RefreshToken: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			},
			err: true,
		}, {
			name: "error_duplicate_uuid",
			arg: pUsers.InsertRefreshTokenParams{
				UserID:       createRandomUser(t).ID,
				RefreshToken: pgtype.UUID{Bytes: refreshToken, Valid: true},
			},
			err: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := repoTest.InsertRefreshToken(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestReadRefreshToken(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	var users []*pUsers.User
	var refreshTokens []uuid.UUID

	for i := 0; i < 3; i++ {
		user := createRandomUser(t)
		users = append(users, user)
		refreshToken := uuid.New()
		refreshTokens = append(refreshTokens, refreshToken)

		inserRefreshTokenTest(t, user.ID, refreshToken)
	}

	testCases := []struct {
		name string
		arg  pUsers.GetRefreshTokenParams
		now  time.Time
		err  bool
	}{
		{
			name: "succes_1",
			arg: pUsers.GetRefreshTokenParams{
				UserID:       users[0].ID,
				RefreshToken: pgtype.UUID{Bytes: refreshTokens[0], Valid: true},
			},
			now: time.Now().UTC(),
			err: false,
		}, {
			name: "succes_2",
			arg: pUsers.GetRefreshTokenParams{
				UserID:       users[1].ID,
				RefreshToken: pgtype.UUID{Bytes: refreshTokens[1], Valid: true},
			},
			now: time.Now().UTC(),
			err: false,
		}, {
			name: "succes_3",
			arg: pUsers.GetRefreshTokenParams{
				UserID:       users[2].ID,
				RefreshToken: pgtype.UUID{Bytes: refreshTokens[2], Valid: true},
			},
			now: time.Now().UTC(),
			err: false,
		}, {
			name: "error_wrong_id",
			arg: pUsers.GetRefreshTokenParams{
				UserID:       users[0].ID + 5,
				RefreshToken: pgtype.UUID{Bytes: refreshTokens[0], Valid: true},
			},
			now: time.Now().UTC(),
			err: true,
		}, {
			name: "error_wrong_uuid",
			arg: pUsers.GetRefreshTokenParams{
				UserID:       users[1].ID,
				RefreshToken: pgtype.UUID{Bytes: refreshTokens[2], Valid: true},
			},
			now: time.Now().UTC(),
			err: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			res, err := repoTest.GetRefreshToken(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, tC.arg.UserID, res.UserID)
				assert.Equal(t, tC.arg.RefreshToken, res.RefreshToken)
				assert.NotZero(t, res.CreatedAt.Time)
				assert.True(t, res.CreatedAt.Valid)
				assert.True(t, res.ExpiresAt.Time.After(tC.now))
				assert.True(t, res.ExpiresAt.Valid)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDeleteRefreshToken(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	var users []*pUsers.User
	for i := 0; i < 3; i++ {
		user := createRandomUser(t)
		users = append(users, user)

		inserRefreshTokenTest(t, user.ID, uuid.New())
	}

	testCases := []struct {
		name   string
		userID int32
		err    bool
	}{
		{
			name:   "succes_1",
			userID: users[0].ID,
			err:    false,
		}, {
			name:   "succes_2",
			userID: users[1].ID,
			err:    false,
		}, {
			name:   "succes_3",
			userID: users[2].ID,
			err:    false,
		}, {
			name:   "error_wrong_id",
			userID: users[0].ID + 5,
			err:    true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err = repoTest.DeleteRefreshToken(ctx, tC.userID)
			if !tC.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUpdateUserIsDeleted(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	var users []pUsers.User
	for i := 0; i < 3; i++ {
		user := createRandomUser(t)
		users = append(users, *user)
	}

	for i := 0; i < 3; i++ {
		arg := pUsers.SoftDeleteUserParams{
			ID:    users[i].ID,
			Email: users[i].Email,
		}
		err = repoTest.SoftDeleteUser(ctx, arg)
		require.NoError(t, err)
	}

	testCases := []struct {
		desc string
		arg  pUsers.CreateUserParams
		user pUsers.User
		err  bool
	}{
		{
			desc: "success",
			arg: pUsers.CreateUserParams{
				Username:       generator.CreateRandomString(5),
				Email:          users[0].Email,
				HashedPassword: generator.CreateRandomString(60),
			},
			user: users[0],
			err:  false,
		}, {
			desc: "success",
			arg: pUsers.CreateUserParams{
				Username:       generator.CreateRandomString(5),
				Email:          users[1].Email,
				HashedPassword: generator.CreateRandomString(60),
			},
			user: users[1],
			err:  false,
		}, {
			desc: "success",
			arg: pUsers.CreateUserParams{
				Username:       generator.CreateRandomString(5),
				Email:          users[2].Email,
				HashedPassword: generator.CreateRandomString(60),
			},
			user: users[2],
			err:  false,
		}, {
			desc: "failed_email_not_found",
			arg: pUsers.CreateUserParams{
				Username:       generator.CreateRandomString(5),
				Email:          generator.CreateRandomEmail(generator.CreateRandomString(5)),
				HashedPassword: generator.CreateRandomString(60),
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.UpdateUserIsDeleted(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, tC.user.ID, res.ID)
				assert.Equal(t, tC.arg.Username, res.Username)
				assert.Equal(t, tC.user.Email, res.Email)
				assert.Equal(t, tC.arg.HashedPassword, res.HashedPassword)
				assert.NotEqual(t, tC.user.CreatedAt, res.CreatedAt)
				assert.False(t, res.IsDeleted.Bool)
				assert.False(t, res.DeletedAt.Time.IsZero())

			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDeleteUserTx(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	var users []*pUsers.User
	for i := 0; i < 4; i++ {
		user := createRandomUser(t)
		users = append(users, user)

		inserRefreshTokenTest(t, user.ID, uuid.New())
	}

	users = append(users, createRandomUser(t))

	testCases := []struct {
		desc string
		arg  pUsers.SoftDeleteUserParams
		err  bool
	}{
		{
			desc: "succes_1",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[0].ID,
				Email: users[0].Email,
			},
			err: false,
		}, {
			desc: "succes_2",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[1].ID,
				Email: users[1].Email,
			},
			err: false,
		}, {
			desc: "succes_3",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[2].ID,
				Email: users[2].Email,
			},
			err: false,
		}, {
			desc: "success_without_refresh_token",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[4].ID,
				Email: users[4].Email,
			},
			err: false,
		}, {
			desc: "error_wrong_id",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[3].ID + 10,
				Email: users[3].Email,
			},
			err: true,
		}, {
			desc: "error_already_deleted",
			arg: pUsers.SoftDeleteUserParams{
				ID:    users[0].ID,
				Email: users[0].Email,
			},
			err: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := repoTest.DeleteUserTx(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
