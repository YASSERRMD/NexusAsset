-- Migration 004: Create servers catalog table.

CREATE TABLE servers (
  id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name            text        NOT NULL,
  hostname        text,
  ip_address      text,
  server_type     text        NOT NULL
                              CHECK (server_type IN ('bare_metal','vm','container','cloud_instance')),
  os              text,
  os_version      text,
  cpu_cores       int,
  ram_gb          int,
  disk_gb         int,
  datacenter      text,
  region          text,
  cloud_provider  text        CHECK (cloud_provider IN ('aws','azure','gcp','on_prem')),
  environment_id  uuid        REFERENCES environments(id),
  owner_team_id   uuid        REFERENCES teams(id),
  managed_by      text        CHECK (managed_by IN ('in_house','vendor_managed')),
  status          text        NOT NULL DEFAULT 'active'
                              CHECK (status IN ('active','decommissioned','maintenance','reserved')),
  tags            text[],
  notes           text,
  created_by      uuid        REFERENCES users(id),
  updated_by      uuid        REFERENCES users(id),
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  deleted_at      timestamptz
);

CREATE INDEX idx_servers_environment ON servers(environment_id);
CREATE INDEX idx_servers_status      ON servers(status);
CREATE INDEX idx_servers_deleted     ON servers(deleted_at) WHERE deleted_at IS NULL;
