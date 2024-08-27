CREATE TABLE metadata (
	version TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	lang TEXT NOT NULL
);

CREATE TABLE asset (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	mime TEXT NOT NULL,
	data BLOB NOT NULL,
	notice TEXT
);

CREATE TABLE step (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT,
	sound INTEGER,
	duration_seconds INTEGER,
	'index' INTEGER UNIQUE ON CONFLICT ABORT
);

CREATE UNIQUE INDEX order_index ON step ('index');

PRAGMA journal_mode = delete;
PRAGMA page_size = 1024;

VACUUM;
