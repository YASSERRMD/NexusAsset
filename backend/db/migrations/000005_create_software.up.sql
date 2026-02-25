-- Migration 005: Create the software catalog (base table only).

CREATE TABLE software (
  id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name           text        NOT NULL,
  display_name   text        NOT NULL,
  description    text,
  version        text,
  software_kind  text        NOT NULL CHECK (software_kind IN ('inhouse','vendor')),
  category_id    uuid        REFERENCES software_categories(id),
  type_id        uuid        REFERENCES software_types(id),
  criticality    text        NOT NULL DEFAULT 'medium'
                             CHECK (criticality IN ('critical','high','medium','low')),
  status         text        NOT NULL DEFAULT 'active'
                             CHECK (status IN ('active','in_development','deprecated','eol','on_hold','decommissioned')),
  architecture   text        CHECK (architecture IN ('monolith','microservice','serverless','library')),
  owner_team_id  uuid        REFERENCES teams(id),
  tags           text[],
  notes          text,
  created_by     uuid        REFERENCES users(id),
  updated_by     uuid        REFERENCES users(id),
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  deleted_at     timestamptz
);

CREATE INDEX idx_software_kind        ON software(software_kind);
CREATE INDEX idx_software_status      ON software(status);
CREATE INDEX idx_software_criticality ON software(criticality);
CREATE INDEX idx_software_category    ON software(category_id);
CREATE INDEX idx_software_deleted     ON software(deleted_at) WHERE deleted_at IS NULL;
