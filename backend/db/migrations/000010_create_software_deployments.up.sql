-- Migration 010: Software deployments (per software per environment).

CREATE TABLE software_deployments (
  id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  software_id      uuid        NOT NULL REFERENCES software(id) ON DELETE CASCADE,
  server_id        uuid        REFERENCES servers(id),
  environment_id   uuid        REFERENCES environments(id),
  version          text,
  deployed_url     text,
  port             int,
  deploy_path      text,
  deployment_type  text        CHECK (deployment_type IN (
                                 'docker','kubernetes','binary','systemd','iis','serverless')),
  status           text        NOT NULL DEFAULT 'running'
                               CHECK (status IN ('running','stopped','failed','deploying','rolled_back')),
  health_status    text        NOT NULL DEFAULT 'unknown'
                               CHECK (health_status IN ('up','down','degraded','unknown')),
  deployed_by      uuid        REFERENCES persons(id),
  deployed_at      timestamptz,
  last_health_check timestamptz,
  notes            text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_depl_software     ON software_deployments(software_id);
CREATE INDEX idx_depl_environment  ON software_deployments(environment_id);
CREATE INDEX idx_depl_health       ON software_deployments(health_status);
