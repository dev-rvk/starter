-- Schema snapshot consumed by sqlc to generate type-safe Go.
-- This file is the single source of truth for the schema in development.

CREATE TABLE IF NOT EXISTS users (
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  id         TEXT PRIMARY KEY,
  username   TEXT NOT NULL UNIQUE,
  email      TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS accounts (
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  id            TEXT PRIMARY KEY,
  email         TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS todo (
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  id          TEXT PRIMARY KEY,
  title       TEXT NOT NULL,
  completed   BOOLEAN DEFAULT FALSE
);
