CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users
(
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    pass_hash BYTEA NOT NULL
);

CREATE TABLE apps
(
    id     INTEGER PRIMARY KEY,
    title   TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);

CREATE TABLE users_tokens
(
    id INTEGER PRIMARY KEY,
    refresh_token TEXT NOT NULL,
    user_id UUID REFERENCES users(id)
);
