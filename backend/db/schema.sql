-- ============================================================
-- CLIENTS AND SITES
-- ============================================================

CREATE TABLE IF NOT EXISTS clients (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    short_code  TEXT NOT NULL UNIQUE,   -- e.g. "ADP", "XYZ"
    notes       TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sites (
    id          BIGSERIAL PRIMARY KEY,
    client_id   BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,          -- e.g. "HQ", "Rome Branch"
    address     TEXT,
    notes       TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(client_id, name)
);

CREATE TABLE IF NOT EXISTS locations (
    id          BIGSERIAL PRIMARY KEY,
    site_id     BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,          -- e.g. "Server Room", "Floor 2", "Reception"
    floor       TEXT,
    notes       TEXT,
    UNIQUE(site_id, name)
);

-- ============================================================
-- IP ADDRESS SPACE
-- ============================================================

-- One block (or more) is assigned to each site.
-- CIDR type enforces valid network notation (e.g. '10.10.0.0/20').
CREATE TABLE IF NOT EXISTS address_blocks (
    id          BIGSERIAL PRIMARY KEY,
    site_id     BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    network     CIDR NOT NULL,          -- e.g. '10.10.0.0/20'
    description TEXT,
    notes       TEXT,
    UNIQUE(site_id, network)
);

-- VLANs are subnets carved from an address_block (e.g. /24 per VLAN).
-- CIDR for subnet, INET for gateway (a host address, not a network address).
CREATE TABLE IF NOT EXISTS vlans (
    id               BIGSERIAL PRIMARY KEY,
    site_id          BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    address_block_id BIGINT REFERENCES address_blocks(id) ON DELETE SET NULL,
    vlan_id          INTEGER NOT NULL,  -- VLAN tag number, e.g. 10, 20, 100
    name             TEXT NOT NULL,     -- e.g. "Users LAN", "VOIP"
    subnet           CIDR,              -- e.g. '10.10.0.0/24'
    gateway          INET,              -- e.g. '10.10.0.1'
    description      TEXT,
    UNIQUE(site_id, vlan_id)
);

-- ============================================================
-- LOOKUP TABLES (manufacturers, categories, suppliers)
-- ============================================================

CREATE TABLE IF NOT EXISTS manufacturers (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,   -- e.g. "HP", "Cisco", "Dell"
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,   -- e.g. "Server", "PC", "Switch", "Printer"
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS suppliers (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    address     TEXT,
    phone       TEXT,
    email       TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- DEVICE CATALOG (defined before switches/devices that reference it)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_models (
    id              BIGSERIAL PRIMARY KEY,
    manufacturer_id BIGINT NOT NULL REFERENCES manufacturers(id),
    model_name      TEXT NOT NULL,      -- e.g. "ProLiant DL360 Gen10"
    category_id     BIGINT NOT NULL REFERENCES categories(id),
    os_default      TEXT,               -- typical OS for this model
    specs           TEXT,               -- free text: CPU, RAM, etc.
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(manufacturer_id, model_name)
);

-- ============================================================
-- NETWORK INFRASTRUCTURE
-- ============================================================

CREATE TABLE IF NOT EXISTS switches (
    id              BIGSERIAL PRIMARY KEY,
    site_id         BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,      -- e.g. "SW006", "CORE-SW01"
    model_id        BIGINT REFERENCES device_models(id) ON DELETE SET NULL,
    ip_address      INET,               -- management IP
    location        TEXT,               -- e.g. "Rack A, Cabinet 3"
    total_ports     INTEGER NOT NULL DEFAULT 24,
    notes           TEXT,
    UNIQUE(site_id, name)
);

CREATE TABLE IF NOT EXISTS switch_ports (
    id              BIGSERIAL PRIMARY KEY,
    switch_id       BIGINT NOT NULL REFERENCES switches(id) ON DELETE CASCADE,
    port_number     INTEGER NOT NULL,
    port_label      TEXT,               -- optional label
    speed           TEXT,               -- e.g. "1G", "10G"
    is_uplink       BOOLEAN DEFAULT FALSE,
    notes           TEXT,
    UNIQUE(switch_id, port_number)
);

CREATE TABLE IF NOT EXISTS patch_panels (
    id              BIGSERIAL PRIMARY KEY,
    site_id         BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,      -- e.g. "PP-RACK1-A"
    total_ports     INTEGER NOT NULL DEFAULT 24,
    location        TEXT,
    notes           TEXT,
    UNIQUE(site_id, name)
);

CREATE TABLE IF NOT EXISTS patch_panel_ports (
    id                  BIGSERIAL PRIMARY KEY,
    patch_panel_id      BIGINT NOT NULL REFERENCES patch_panels(id) ON DELETE CASCADE,
    port_number         INTEGER NOT NULL,
    port_label          TEXT,
    notes               TEXT,
    UNIQUE(patch_panel_id, port_number)
);

-- ============================================================
-- DEPLOYED DEVICES
-- ============================================================

CREATE TABLE IF NOT EXISTS devices (
    id                  BIGSERIAL PRIMARY KEY,
    site_id             BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    location_id         BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    model_id            BIGINT REFERENCES device_models(id) ON DELETE SET NULL,

    -- Identification
    hostname            TEXT NOT NULL,
    dns_name            TEXT,
    serial_number       TEXT,
    asset_tag           TEXT,

    -- Type and status
    category_id         BIGINT NOT NULL REFERENCES categories(id),
    status              TEXT NOT NULL DEFAULT 'active',  -- active, inactive, reserved, decommissioned
    is_up               BOOLEAN DEFAULT TRUE,

    -- Software / management
    os                  TEXT,
    has_rmm             BOOLEAN DEFAULT FALSE,  -- RMM agent installed
    has_antivirus       BOOLEAN DEFAULT FALSE,  -- antivirus installed
    supplier_id         BIGINT REFERENCES suppliers(id) ON DELETE SET NULL,

    -- Logistics
    installation_date   DATE,
    is_reserved         BOOLEAN DEFAULT FALSE,

    notes               TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, hostname)
);

-- Trigger function to auto-update updated_at on any UPDATE.
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS devices_updated_at ON devices;
CREATE TRIGGER devices_updated_at
BEFORE UPDATE ON devices
FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- NETWORK INTERFACES (NICs on a device)
-- ============================================================

-- Each row is one physical or virtual NIC on a device.
-- A PC typically has one; a server may have eth0, eth1, iDRAC;
-- a router may have WAN, LAN1, LAN2, mgmt.
CREATE TABLE IF NOT EXISTS device_interfaces (
    id          BIGSERIAL PRIMARY KEY,
    device_id   BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,      -- e.g. "eth0", "iDRAC", "WAN", "LAN1"
    mac_address MACADDR,            -- MAC address of this NIC
    notes       TEXT,
    UNIQUE(device_id, name)
);

-- ============================================================
-- IP ADDRESSES (one or more IPs per NIC)
-- ============================================================

-- INET type stores a host address with optional prefix (e.g. '10.0.0.1/24').
CREATE TABLE IF NOT EXISTS device_ips (
    id              BIGSERIAL PRIMARY KEY,
    interface_id    BIGINT NOT NULL REFERENCES device_interfaces(id) ON DELETE CASCADE,
    ip_address      INET NOT NULL,
    vlan_id         BIGINT REFERENCES vlans(id),
    is_primary      BOOLEAN DEFAULT FALSE,  -- primary IP of the whole device
    notes           TEXT
);

-- ============================================================
-- PHYSICAL CONNECTIONS (NIC → switch port / patch panel port)
-- ============================================================

-- Tracks which NIC is physically plugged into which switch port
-- and/or patch panel port. Both columns are optional because a
-- cable may bypass the patch panel and go directly to the switch.
CREATE TABLE IF NOT EXISTS device_connections (
    id                  BIGSERIAL PRIMARY KEY,
    interface_id        BIGINT NOT NULL REFERENCES device_interfaces(id) ON DELETE CASCADE,
    switch_port_id      BIGINT REFERENCES switch_ports(id),
    patch_panel_port_id BIGINT REFERENCES patch_panel_ports(id),
    connected_at        DATE,
    notes               TEXT
);

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_devices_site             ON devices(site_id);
CREATE INDEX IF NOT EXISTS idx_devices_hostname         ON devices(hostname);
CREATE INDEX IF NOT EXISTS idx_device_interfaces_device ON device_interfaces(device_id);
CREATE INDEX IF NOT EXISTS idx_device_ips_address       ON device_ips(ip_address);
CREATE INDEX IF NOT EXISTS idx_switch_ports_sw          ON switch_ports(switch_id);
CREATE INDEX IF NOT EXISTS idx_address_blocks_site      ON address_blocks(site_id);
CREATE INDEX IF NOT EXISTS idx_vlans_block              ON vlans(address_block_id);
