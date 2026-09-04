CREATE TABLE IF NOT EXISTS users (
	id				TEXT	PRIMARY KEY,
	username		TEXT	NOT NULL UNIQUE,
    password_hash   TEXT    NOT NULL,
	sleeper_user_id	TEXT,
	espn_swid		TEXT,
	espn_s2			TEXT,
	created_at		DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token       TEXT        PRIMARY KEY,
    user_id     TEXT        NOT NULL,
    expires_at  DATETIME    NOT NULL,
    created_at  DATETIME    DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS leagues (
	id					TEXT	PRIMARY KEY,
	user_id				TEXT	NOT NULL,
	platform			TEXT	NOT NULL,
	external_league_id	TEXT	NOT NULL,
	name				TEXT	NOT NULL,
	season				TEXT	NOT NULL,
	created_at			DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, external_league_id)
);