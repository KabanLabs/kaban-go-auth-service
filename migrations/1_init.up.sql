CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    pass_hash BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE apps
(
    id     SERIAL PRIMARY KEY,
    title   TEXT NOT NULL UNIQUE,
    scopes TEXT
);

CREATE TABLE users_tokens
(
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL,
    app_id INTEGER NOT NULL REFERENCES apps(id),
    user_id UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    rotated BOOLEAN DEFAULT FALSE
);
