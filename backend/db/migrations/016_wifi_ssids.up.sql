CREATE TABLE IF NOT EXISTS wifi_ssids (
    id       BIGSERIAL PRIMARY KEY,
    site_id  BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    ssid     TEXT NOT NULL,
    auth     TEXT,
    vlan_id  BIGINT REFERENCES vlans(id) ON DELETE SET NULL,
    notes    TEXT,
    UNIQUE(site_id, ssid)
);
