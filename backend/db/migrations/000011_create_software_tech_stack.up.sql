-- Migration 011: Software tech stack and database links.

CREATE TABLE software_tech_stack (
  id               uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  software_id      uuid  NOT NULL REFERENCES software(id) ON DELETE CASCADE,
  technology       text  NOT NULL,
  tech_category_id uuid  REFERENCES tech_categories(id),
  version          text
);

CREATE INDEX idx_ts_software ON software_tech_stack(software_id);

-- software_database_links is here too (depends on database_instances in 013)
-- This table is created in migration 013 instead to keep FK ordering correct.
