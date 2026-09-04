-- name: GetUser :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: CreateUser :exec
INSERT INTO users (id, username, password_hash, sleeper_user_id, espn_swid, espn_s2)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetLeaguesByUser :many
SELECT * FROM leagues
WHERE user_id = ?
ORDER BY name ASC;

-- name: CreateLeague :exec
INSERT INTO leagues (id, user_id, platform, external_league_id, name, season)
VALUES (?, ?, ?, ?, ?, ?);

-- name: DeleteLeague :exec
DELETE FROM leagues
WHERE id = ?;

-- name: CreateSession :exec
INSERT INTO sessions (token, user_id, expires_at)
VALUES (?, ?, ?);

-- name: GetUserBySessionToken :one
SELECT users.* FROM users
JOIN sessions ON users.id = sessions.user_id
WHERE sessions.token = ? AND sessions.expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = ?;

-- name: GetSession :one
SELECT * FROM sessions
WHERE token = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;