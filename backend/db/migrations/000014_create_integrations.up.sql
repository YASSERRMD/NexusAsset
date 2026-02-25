-- Migration 014: Integration catalog (exposed + consumed APIs).

CREATE TABLE integrations_exposed (
  id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  software_id         uuid        NOT NULL REFERENCES software(id) ON DELETE CASCADE,
  name                text        NOT NULL,
  description         text,
  protocol            text        NOT NULL
                                  CHECK (protocol IN ('rest','soap','grpc','graphql',
                                    'mqtt','amqp','kafka','ftp','sftp','websocket')),
  endpoint_url        text,
  port                int,
  auth_method         text        CHECK (auth_method IN ('api_key','oauth2','mtls','basic','jwt','none')),
  version             text,
  spec_url            text,                    -- OpenAPI/WSDL/proto link
  sla_uptime_percent  numeric,
  owner_team_id       uuid        REFERENCES teams(id),
  status              text        NOT NULL DEFAULT 'active',
  deprecation_date    date,
  notes               text,
  created_by          uuid        REFERENCES users(id),
  updated_by          uuid        REFERENCES users(id),
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  deleted_at          timestamptz
);

CREATE TABLE integration_consumers (
  id               uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  integration_id   uuid  NOT NULL REFERENCES integrations_exposed(id) ON DELETE CASCADE,
  consumer_name    text  NOT NULL,
  consumer_type    text  CHECK (consumer_type IN ('internal','external')),
  contact_email    text,
  notes            text
);

CREATE TABLE integrations_consumed (
  id                      uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  software_id             uuid        NOT NULL REFERENCES software(id) ON DELETE CASCADE,
  name                    text        NOT NULL,
  description             text,
  protocol                text        NOT NULL
                                      CHECK (protocol IN ('rest','soap','grpc','graphql',
                                        'mqtt','amqp','kafka','ftp','sftp','websocket')),
  vendor_id               uuid        REFERENCES vendors(id),
  endpoint_url            text,
  auth_method             text        CHECK (auth_method IN ('api_key','oauth2','mtls','basic','jwt','none')),
  dependency_criticality  text        CHECK (dependency_criticality IN ('critical','high','medium','low')),
  fallback_strategy       text,
  status                  text        NOT NULL DEFAULT 'active',
  notes                   text,
  created_by              uuid        REFERENCES users(id),
  updated_by              uuid        REFERENCES users(id),
  created_at              timestamptz NOT NULL DEFAULT now(),
  updated_at              timestamptz NOT NULL DEFAULT now(),
  deleted_at              timestamptz
);

CREATE INDEX idx_iexp_software   ON integrations_exposed(software_id);
CREATE INDEX idx_iexp_deleted    ON integrations_exposed(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_icons_software  ON integrations_consumed(software_id);
CREATE INDEX idx_icons_deleted   ON integrations_consumed(deleted_at) WHERE deleted_at IS NULL;
