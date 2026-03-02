PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

-- ============================================================
-- CLIENTS AND SITES
-- ============================================================

CREATE TABLE IF NOT EXISTS clients (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    short_code  TEXT NOT NULL UNIQUE,   -- e.g. "ADP", "XYZ"
    notes       TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sites (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id   INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,          -- e.g. "HQ", "Rome Branch"
    address     TEXT,
    notes       TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(client_id, name)
);

CREATE TABLE IF NOT EXISTS offices (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id     INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,          -- e.g. "IT Dept", "Reception"
    floor       TEXT,
    notes       TEXT,
    UNIQUE(site_id, name)
);

-- ============================================================
-- IP ADDRESS SPACE
-- ============================================================

-- One /20 block (or any prefix) is assigned to each site.
-- Multiple blocks per site are allowed for flexibility.
CREATE TABLE IF NOT EXISTS address_blocks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id     INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    network     TEXT NOT NULL,          -- e.g. "10.10.0.0/20"
    description TEXT,
    notes       TEXT,
    UNIQUE(site_id, network)
);

-- VLANs are subnets carved from an address_block (e.g. /24 per VLAN)
CREATE TABLE IF NOT EXISTS vlans (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id          INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    address_block_id INTEGER REFERENCES address_blocks(id),
    vlan_id          INTEGER NOT NULL,  -- VLAN number, e.g. 10, 20, 100
    name             TEXT NOT NULL,     -- e.g. "Users LAN", "VOIP"
    subnet           TEXT,              -- e.g. "10.10.0.0/24"
    gateway          TEXT,
    description      TEXT,
    UNIQUE(site_id, vlan_id)
);

-- ============================================================
-- NETWORK INFRASTRUCTURE
-- ============================================================

CREATE TABLE IF NOT EXISTS switches (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id         INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,      -- e.g. "SW006", "CORE-SW01"
    model_id        INTEGER REFERENCES device_models(id),
    ip_address      TEXT,
    location        TEXT,               -- e.g. "Rack A, Cabinet 3"
    total_ports     INTEGER NOT NULL DEFAULT 24,
    notes           TEXT,
    UNIQUE(site_id, name)
);

CREATE TABLE IF NOT EXISTS switch_ports (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    switch_id       INTEGER NOT NULL REFERENCES switches(id) ON DELETE CASCADE,
    port_number     INTEGER NOT NULL,
    port_label      TEXT,               -- optional label
    speed           TEXT,               -- e.g. "1G", "10G"
    is_uplink       BOOLEAN DEFAULT 0,
    notes           TEXT,
    UNIQUE(switch_id, port_number)
);

CREATE TABLE IF NOT EXISTS patch_panels (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id         INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,      -- e.g. "PP-RACK1-A"
    total_ports     INTEGER NOT NULL DEFAULT 24,
    location        TEXT,
    notes           TEXT,
    UNIQUE(site_id, name)
);

CREATE TABLE IF NOT EXISTS patch_panel_ports (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    patch_panel_id      INTEGER NOT NULL REFERENCES patch_panels(id) ON DELETE CASCADE,
    port_number         INTEGER NOT NULL,
    port_label          TEXT,
    notes               TEXT,
    UNIQUE(patch_panel_id, port_number)
);

-- ============================================================
-- DEVICE CATALOG (INVENTORY)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_models (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    manufacturer    TEXT NOT NULL,      -- e.g. "HP", "Cisco", "Dell"
    model_name      TEXT NOT NULL,      -- e.g. "ProLiant DL360 Gen10"
    category        TEXT NOT NULL,      -- Server, PC, Laptop, Printer, Switch, Router, AP, NAS, Camera, Phone, UPS, Other
    os_default      TEXT,               -- typical OS for this model
    specs           TEXT,               -- free text: CPU, RAM, etc.
    notes           TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(manufacturer, model_name)
);

-- ============================================================
-- DEPLOYED DEVICES
-- ============================================================

CREATE TABLE IF NOT EXISTS devices (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id             INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    office_id           INTEGER REFERENCES offices(id),
    model_id            INTEGER REFERENCES device_models(id),

    -- Identification
    hostname            TEXT NOT NULL,
    dns_name            TEXT,
    serial_number       TEXT,
    asset_tag           TEXT,

    -- Type and status
    device_type         TEXT NOT NULL,  -- PC, Server, Printer, Switch, AP, Camera, Phone, NAS, UPS, Other
    status              TEXT NOT NULL DEFAULT 'active',  -- active, inactive, reserved, decommissioned
    is_up               BOOLEAN DEFAULT 1,

    -- Software / management
    os                  TEXT,
    has_rmm             BOOLEAN DEFAULT 0,  -- RMM agent installed
    has_antivirus       BOOLEAN DEFAULT 0,  -- antivirus installed
    supplier            TEXT,

    -- Logistics
    installation_date   DATE,
    is_reserved         BOOLEAN DEFAULT 0,

    -- Ticket / reason
    ticket_ref          TEXT,
    reason              TEXT,

    notes               TEXT,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Trigger to auto-update updated_at
CREATE TRIGGER IF NOT EXISTS devices_updated_at
AFTER UPDATE ON devices
BEGIN
    UPDATE devices SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- ============================================================
-- IP ADDRESSES (multiple IPs per device)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_ips (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id   INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    ip_address  TEXT NOT NULL,
    mac_address TEXT,
    vlan_id     INTEGER REFERENCES vlans(id),
    is_primary  BOOLEAN DEFAULT 0,
    interface   TEXT,                   -- e.g. "eth0", "Wi-Fi"
    notes       TEXT
);

-- ============================================================
-- PHYSICAL CONNECTIONS (device → switch port / patch panel port)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_connections (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id           INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    switch_port_id      INTEGER REFERENCES switch_ports(id),
    patch_panel_port_id INTEGER REFERENCES patch_panel_ports(id),
    connected_at        DATE,
    notes               TEXT
);

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_devices_site         ON devices(site_id);
CREATE INDEX IF NOT EXISTS idx_devices_hostname     ON devices(hostname);
CREATE INDEX IF NOT EXISTS idx_device_ips_address   ON device_ips(ip_address);
CREATE INDEX IF NOT EXISTS idx_switch_ports_sw      ON switch_ports(switch_id);
CREATE INDEX IF NOT EXISTS idx_address_blocks_site  ON address_blocks(site_id);
CREATE INDEX IF NOT EXISTS idx_vlans_block          ON vlans(address_block_id);
