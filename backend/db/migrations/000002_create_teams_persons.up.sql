-- Migration 002: Create teams and persons tables.
-- teams and persons have a circular FK (team.manager_id → persons),
-- so we first create without the FK constraint then add it after.

CREATE TABLE teams (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text        NOT NULL,
  department  text        NOT NULL,
  team_email  text,
  manager_id  uuid,                        -- FK added after persons is created
  is_active   bool        NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
);

CREATE TABLE persons (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name   text        NOT NULL,
  email       text        NOT NULL UNIQUE,
  phone       text,
  title       text,
  department  text,
  team_id     uuid        REFERENCES teams(id),
  is_active   bool        NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
);

-- Add the circular FK now that persons exists
ALTER TABLE teams
  ADD CONSTRAINT fk_teams_manager
  FOREIGN KEY (manager_id) REFERENCES persons(id);
