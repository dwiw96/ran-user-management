package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"os"
	"testing"
	"time"

	pUsers "github.com/dwiw96/ran-user-management/internal/features/users"
	generator "github.com/dwiw96/ran-user-management/pkg/utils/generator"
	password "github.com/dwiw96/ran-user-management/pkg/utils/password"
	testUtils "github.com/dwiw96/ran-user-management/testutils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	poolTest   *pgxpool.Pool
	ctxTest    context.Context
	clientTest *redis.Client
)

func TestMain(m *testing.M) {
	poolTest = testUtils.GetPool()
	defer poolTest.Close()
	ctxTest = testUtils.GetContext()
	defer ctxTest.Done()
	clientTest = testUtils.GetRedisClient()
	defer clientTest.Close()

	schemaCleanup := testUtils.SetupDB("test_middleware_auth")

	password.JwtInit(poolTest, ctxTest)

	var cancel context.CancelFunc
	ctxTest, cancel = context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	exitTest := m.Run()

	schemaCleanup()

	os.Exit(exitTest)
}

func createTokenAndKey(t *testing.T) (string, pUsers.User, *rsa.PrivateKey) {
	key, err := LoadKey(ctxTest, poolTest)
	require.NoError(t, err)
	require.NotNil(t, key)

	firstname := generator.CreateRandomString(5)
	payload := pUsers.User{
		Username: firstname + " " + generator.CreateRandomString(7),
		Email:    generator.CreateRandomEmail(firstname),
	}

	token, err := CreateToken(payload, 5, key)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	return token, payload, key
}

func TestCreateToken(t *testing.T) {
	createTokenAndKey(t)
}

func TestVerifyToken(t *testing.T) {
	token, _, key := createTokenAndKey(t)

	t.Run("success", func(t *testing.T) {
		res, err := VerifyToken(token, key)
		require.NoError(t, err)
		require.True(t, res)
	})

	t.Run("failed", func(t *testing.T) {
		res, err := VerifyToken(token+"b", key)
		require.Error(t, err)
		require.False(t, res)
	})
}

func TestReadToken(t *testing.T) {
	token, payloadInput, key := createTokenAndKey(t)

	t.Run("success", func(t *testing.T) {
		payload, err := ReadToken(token, key)
		require.NoError(t, err)
		assert.Equal(t, payloadInput.Username, payload.Name)
		assert.Equal(t, payloadInput.Email, payload.Email)
	})

	t.Run("failed", func(t *testing.T) {
		payload, err := ReadToken(token+"b", key)
		require.Error(t, err)
		assert.Nil(t, payload)
	})
}

func TestLoadKey(t *testing.T) {
	res, err := LoadKey(ctxTest, poolTest)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestCheckBlockedToken(t *testing.T) {
	token, _, key := createTokenAndKey(t)

	payload, err := ReadToken(token, key)
	require.NoError(t, err)

	t.Run("valid", func(t *testing.T) {
		err = CheckBlockedToken(clientTest, ctxTest, payload.ID)
		require.NoError(t, err)
	})

	t.Run("blacklist", func(t *testing.T) {
		iat := time.Unix(payload.Iat, 0)
		exp := time.Unix(payload.Exp, 0)
		duration := time.Duration(exp.Sub(iat).Nanoseconds())
		err = clientTest.Set(ctxTest, "block "+payload.ID.String(), payload.UserID, duration).Err()
		require.NoError(t, err)

		err = CheckBlockedToken(clientTest, ctxTest, payload.ID)
		require.Error(t, err)
	})
}

func TestGetHeaderToken(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", nil)
	require.NoError(t, err)

	authHeader := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImYyNGZiYmExLTE5NDctNGNhYy05ODA4LTM2ZDY2YzQ2NzIwMCIsInVzZXJfaWQiOjc2LCJpc3MiOiIiLCJuYW1lIjoiR3JhY2UgRG9lIEp1bmlvciIsImVtYWlsIjoiZ3JhY2VAbWFpbC5jb20iLCJhZGRyZXNzIjoiQ2lyY2xlIFN0cmVldCwgTm8uMSwgQmFuZHVuZywgV2VzdCBKYXZhIiwiaWF0IjoxNzI3ODM3MjY5LCJleHAiOjE3Mjc4NDA4Njl9.mdQtJ22xRT5n8xYp5dGdVIzBo-OOocnaE6F054C0LEImf1rA_Fo0_fd3IGVa3XW5kDdpobqB8K6hDFm-XCPbkxvIfXjsjAwGqDrlzsjLiNmSvRwUj6FFWUkIpS_4Nl7Szcc2dEXe7n75LOs9yIhzNmuNjyC9Ago8BJiTYL0_jAkzxlHUwSaRj6naxbsLpiRhpjAW14-ema0wdbbHkaPkv0cj6rOQlsRTCW6R6i_2lrew5eOHIR750gBdImJ8HGtzB29yUA3A9P0-rGjITwZTanoqtOdv5d6lSMJ7eYMEACe4Lj3-k93V65e2ZJEFCnutk0H2ZPSaMBZwTx9B32S8JQ"

	r.Header.Set("Authorization", "Bearer "+authHeader)

	token, err := GetTokenHeader(r)
	require.NoError(t, err)
	assert.Equal(t, authHeader, token)
}

func TestPayloadVerification(t *testing.T) {
	var user pUsers.User
	user.Username = generator.CreateRandomString(int(generator.RandomInt(3, 13)))
	user.Email = generator.CreateRandomEmail(user.Username)
	user.HashedPassword = generator.CreateRandomString(int(generator.RandomInt(5, 10)))

	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.HashedPassword)

	query := `
	INSERT INTO users(
		email,
		username,
		hashed_password
	) VALUES 
		($1, $2, $3) 
	RETURNING id;`

	row := poolTest.QueryRow(ctxTest, query, user.Email, user.Username, user.HashedPassword)
	err := row.Scan(&user.ID)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)

	err = PayloadVerification(ctxTest, poolTest, user.Email, user.Username)
	require.NoError(t, err)
}
