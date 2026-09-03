-- name: GetUser :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: CreateUser :exec
INSERT INTO users (id, username, sleeper_user_id, espn_swid, espn_s2)
VALUES (?, ?, ?, ?, ?);

-- name: GetLeaguesByUser :many
SELECT * FROM leagues
WHERE user_id = ?
ORDER BY name ASC;

-- name: CreateLeague :exec
INSERT INTO leagues (id, user_id, platform, external_league_id, name, season)
VALUES (?, ?, ?, ?, ?, ?);

-- name: DeleteLeague :exec
DELETE FROM leagues
WHERE id = ?