-- Migration 008: Software responsibilities (RACI-style, multi-person).

CREATE TABLE software_responsibilities (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  software_id uuid        NOT NULL REFERENCES software(id) ON DELETE CASCADE,
  person_id   uuid        NOT NULL REFERENCES persons(id),
  role_id     uuid        NOT NULL REFERENCES responsibility_roles(id),
  is_primary  bool        NOT NULL DEFAULT false,
  assigned_at timestamptz NOT NULL DEFAULT now(),
  notes       text,
  UNIQUE (software_id, person_id, role_id)
);

CREATE INDEX idx_sr_software ON software_responsibilities(software_id);
CREATE INDEX idx_sr_person   ON software_responsibilities(person_id);
