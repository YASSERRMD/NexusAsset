-- Migration 003: Create users table for authentication.
-- role column enforces 'admin' | 'contributor' | 'reader'

CREATE TABLE users (
  id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  person_id      uuid        REFERENCES persons(id),
  username       text        NOT NULL UNIQUE,
  email          text        NOT NULL UNIQUE,
  password_hash  text        NOT NULL,
  role           text        NOT NULL DEFAULT 'reader'
                             CHECK (role IN ('admin', 'contributor', 'reader')),
  is_active      bool        NOT NULL DEFAULT true,
  last_login_at  timestamptz,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role     ON users(role);
