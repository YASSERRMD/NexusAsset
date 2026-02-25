-- Migration 012: Vendor catalog and contracts.

CREATE TABLE vendors (
  id                     uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name                   text        NOT NULL,
  country                text,
  vendor_type            text        CHECK (vendor_type IN ('software','hardware','cloud','support','consulting')),
  website                text,
  primary_contact_name   text,
  primary_contact_email  text,
  primary_contact_phone  text,
  support_email          text,
  support_phone          text,
  notes                  text,
  tags                   text[],
  status                 text        NOT NULL DEFAULT 'active',
  created_by             uuid        REFERENCES users(id),
  updated_by             uuid        REFERENCES users(id),
  created_at             timestamptz NOT NULL DEFAULT now(),
  updated_at             timestamptz NOT NULL DEFAULT now(),
  deleted_at             timestamptz
);

CREATE TABLE vendor_contracts (
  id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  vendor_id       uuid        NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
  contract_ref    text,
  scope           text,
  start_date      date,
  end_date        date,
  sla_terms       text,
  auto_renew      bool        NOT NULL DEFAULT false,
  value_amount    numeric,
  value_currency  text        NOT NULL DEFAULT 'USD',
  document_url    text,
  notes           text,
  created_at      timestamptz NOT NULL DEFAULT now()
);

-- Now add the FK from software_vendor_details to vendors
ALTER TABLE software_vendor_details
  ADD CONSTRAINT fk_svd_vendor
  FOREIGN KEY (vendor_id) REFERENCES vendors(id);

CREATE INDEX idx_vendors_deleted   ON vendors(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_contracts_vendor  ON vendor_contracts(vendor_id);
CREATE INDEX idx_contracts_expiry  ON vendor_contracts(end_date);
