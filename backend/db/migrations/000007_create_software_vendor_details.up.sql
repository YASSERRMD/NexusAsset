-- Migration 007: Vendor software extension details.
-- vendor_id FK to vendors is added in migration 012 after vendors table is created.

CREATE TABLE software_vendor_details (
  software_id                uuid  PRIMARY KEY REFERENCES software(id) ON DELETE CASCADE,
  vendor_id                  uuid,          -- FK → vendors (added via migration 012)
  license_type               text,
  license_key                text,
  license_expiry_date        date,
  support_contract_ref       text,
  installed_version          text,
  latest_available_version   text,
  eol_date                   date,
  purchase_date              date
);

CREATE INDEX idx_svd_vendor ON software_vendor_details(vendor_id);
CREATE INDEX idx_svd_license_expiry ON software_vendor_details(license_expiry_date);
