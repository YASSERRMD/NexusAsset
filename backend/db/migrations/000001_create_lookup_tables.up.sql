-- Migration 001: Create all lookup tables
-- These are admin-managed extensible reference tables.

-- +migrate Up

CREATE TABLE software_categories (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text        NOT NULL UNIQUE,
  description text,
  color       text,                        -- hex color for UI badge e.g. #3B82F6
  is_active   bool        NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE software_types (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text        NOT NULL UNIQUE, -- 'Web App','REST API','Batch Job',etc.
  description text,
  icon        text,                        -- lucide icon name
  is_active   bool        NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE responsibility_roles (
  id          uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text  NOT NULL UNIQUE, -- 'product_owner','tech_lead','dev_lead', etc.
  description text,
  is_active   bool  NOT NULL DEFAULT true
);

CREATE TABLE environments (
  id           uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  name         text  NOT NULL UNIQUE,  -- 'dev','qa','staging','production', etc.
  display_name text  NOT NULL,
  color        text,                   -- badge color
  order_index  int   NOT NULL DEFAULT 0,
  is_active    bool  NOT NULL DEFAULT true
);

CREATE TABLE repo_platforms (
  id           uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  name         text  NOT NULL UNIQUE, -- 'github','gitlab','bitbucket', etc.
  display_name text  NOT NULL,
  icon         text,
  is_active    bool  NOT NULL DEFAULT true
);

CREATE TABLE tech_categories (
  id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL UNIQUE  -- 'language','framework','database', etc.
);
