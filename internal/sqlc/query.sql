-- name: CreateUser :one
INSERT INTO users(
    username,
    email,
    hashed_password
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND is_deleted = FALSE;

-- name: UpdateUser :one
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
RETURNING *;

-- name: SoftDeleteUser :exec
UPDATE
    users
SET
    is_deleted = TRUE,
    deleted_at = NOW()
WHERE
    id = $1
AND email = $2;

-- name: InsertRefreshToken :exec
INSERT INTO 
    refresh_token_whitelist(user_id, refresh_token, expires_at) 
VALUES(
    $1, $2, NOW() + INTERVAL '5 minute'
);

-- name: GetRefreshToken :one
SELECT * FROM refresh_token_whitelist WHERE user_id = $1 AND refresh_token = $2;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_token_whitelist WHERE user_id = $1;

-- name: UpdateUserIsDeleted :one
UPDATE
    users
SET
    is_deleted = FALSE
WHERE
    email = $1
RETURNING *;
