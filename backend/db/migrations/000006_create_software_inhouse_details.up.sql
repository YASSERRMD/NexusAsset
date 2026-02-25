-- Migration 006: In-house software extension details.

CREATE TABLE software_inhouse_details (
  software_id         uuid        PRIMARY KEY REFERENCES software(id) ON DELETE CASCADE,
  cicd_platform       text,
  cicd_pipeline_url   text,
  last_released_at    timestamptz,
  next_release_at     timestamptz,
  documentation_url   text,
  internal_notes      text
);
