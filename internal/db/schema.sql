CREATE TABLE IF NOT EXISTS users (
	id				TEXT	PRIMARY KEY,
	username		TEXT	NOT NULL UNIQUE,
	sleeper_user_id	TEXT,
	espn_swid		TEXT,
	espn_s2			TEXT,
	craeted_at		DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS leagues (
	id					TEXT	PRIMARY KEY,
	user_id				TEXT	NOT NULL,
	platform			TEXT	NOT NULL,
	external_league_id	TEXT	NOT NULL,
	name				TEXT	NOT NULL,
	season				TEXT	NOT NULL,
	craeted_at			DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);