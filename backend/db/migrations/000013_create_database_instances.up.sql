-- Migration 013: Database instances catalog + software-database links.

CREATE TABLE database_instances (
  id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name             text        NOT NULL,
  db_type          text        NOT NULL
                               CHECK (db_type IN ('postgresql','mysql','oracle','mssql','mongodb',
                                 'redis','cassandra','elasticsearch','influxdb','neo4j')),
  version          text,
  host_server_id   uuid        REFERENCES servers(id),
  environment_id   uuid        REFERENCES environments(id),
  port             int,
  size_gb          numeric,
  owner_team_id    uuid        REFERENCES teams(id),
  classification   text        CHECK (classification IN ('public','internal','confidential','restricted')),
  backup_policy    text,
  status           text        NOT NULL DEFAULT 'active',
  notes            text,
  created_by       uuid        REFERENCES users(id),
  updated_by       uuid        REFERENCES users(id),
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  deleted_at       timestamptz
);

CREATE TABLE software_database_links (
  id                   uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  software_id          uuid  NOT NULL REFERENCES software(id) ON DELETE CASCADE,
  database_instance_id uuid  NOT NULL REFERENCES database_instances(id),
  access_type          text  NOT NULL
                             CHECK (access_type IN ('read','write','read_write','admin')),
  schema_name          text,
  notes                text
);

CREATE INDEX idx_db_environment ON database_instances(environment_id);
CREATE INDEX idx_db_deleted     ON database_instances(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_dblink_sw      ON software_database_links(software_id);
CREATE INDEX idx_dblink_db      ON software_database_links(database_instance_id);
