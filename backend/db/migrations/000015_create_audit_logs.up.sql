-- Migration 015: Audit log table.

CREATE TABLE audit_logs (
  id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      uuid        REFERENCES users(id),
  entity_type  text        NOT NULL,   -- 'software','server','vendor','database', etc.
  entity_id    uuid        NOT NULL,
  action       text        NOT NULL CHECK (action IN ('create','update','delete','view','export')),
  changes      jsonb,                  -- {field: {old: x, new: y}}
  ip_address   text,
  user_agent   text,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_user        ON audit_logs(user_id);
CREATE INDEX idx_audit_entity      ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_action      ON audit_logs(action);
CREATE INDEX idx_audit_created_at  ON audit_logs(created_at DESC);
