-- Migration 009: Software repositories (multi-repo per software).

CREATE TABLE software_repositories (
  id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  software_id     uuid        NOT NULL REFERENCES software(id) ON DELETE CASCADE,
  name            text        NOT NULL,           -- 'backend', 'frontend', 'infra'
  repo_type       text        NOT NULL
                              CHECK (repo_type IN ('backend','frontend','mobile',
                                'infrastructure','shared_library','ml_model','configuration')),
  url             text        NOT NULL,
  default_branch  text        NOT NULL DEFAULT 'main',
  platform_id     uuid        REFERENCES repo_platforms(id),
  is_private      bool        NOT NULL DEFAULT true,
  last_commit_at  timestamptz,
  primary_dev_id  uuid        REFERENCES persons(id),
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_repo_software  ON software_repositories(software_id);
CREATE INDEX idx_repo_platform  ON software_repositories(platform_id);
