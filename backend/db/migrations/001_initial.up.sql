-- ============================================================
-- USERS AND AUTHENTICATION
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_admin      BOOLEAN DEFAULT false,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- CLIENTS AND SITES
-- ============================================================

CREATE TABLE IF NOT EXISTS clients (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    short_code  TEXT NOT NULL UNIQUE,
    domain      TEXT,
    notes       TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sites (
    id          BIGSERIAL PRIMARY KEY,
    client_id   BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    address     TEXT,
    domain      TEXT,
    notes       TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(client_id, name)
);

CREATE TABLE IF NOT EXISTS locations (
    id          BIGSERIAL PRIMARY KEY,
    site_id     BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    floor       TEXT,
    notes       TEXT,
    UNIQUE(site_id, name)
);

-- ============================================================
-- IP ADDRESS SPACE
-- ============================================================

CREATE TABLE IF NOT EXISTS address_blocks (
    id          BIGSERIAL PRIMARY KEY,
    site_id     BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    network     CIDR NOT NULL,
    description TEXT,
    notes       TEXT,
    UNIQUE(site_id, network)
);

CREATE TABLE IF NOT EXISTS vlans (
    id               BIGSERIAL PRIMARY KEY,
    site_id          BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    address_block_id BIGINT REFERENCES address_blocks(id) ON DELETE SET NULL,
    vlan_id          INTEGER NOT NULL,
    name             TEXT NOT NULL,
    subnet           CIDR,
    gateway_device_ip_id BIGINT,
    dhcp_start       INET,
    dhcp_end         INET,
    description      TEXT,
    UNIQUE(site_id, vlan_id)
);

-- ============================================================
-- LOOKUP TABLES
-- ============================================================

CREATE TABLE IF NOT EXISTS manufacturers (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    short_code  TEXT NOT NULL,
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

CREATE TABLE IF NOT EXISTS operating_systems (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- DEVICE CATALOG
-- ============================================================

CREATE TABLE IF NOT EXISTS device_models (
    id              BIGSERIAL PRIMARY KEY,
    manufacturer_id BIGINT NOT NULL REFERENCES manufacturers(id),
    model_name      TEXT NOT NULL,
    category_id     BIGINT NOT NULL REFERENCES categories(id),
    os_default_id   BIGINT REFERENCES operating_systems(id) ON DELETE SET NULL,
    default_ports   INTEGER,
    specs           TEXT,
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
    hostname        TEXT NOT NULL,
    model_id        BIGINT REFERENCES device_models(id) ON DELETE SET NULL,
    ip_address      INET,
    vlan_id         BIGINT REFERENCES vlans(id) ON DELETE SET NULL,
    location_id     BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    total_ports     INTEGER NOT NULL DEFAULT 24,
    notes           TEXT,
    UNIQUE(site_id, hostname)
);

CREATE TABLE IF NOT EXISTS switch_ports (
    id              BIGSERIAL PRIMARY KEY,
    switch_id       BIGINT NOT NULL REFERENCES switches(id) ON DELETE CASCADE,
    port_number     INTEGER NOT NULL,
    port_label      TEXT,
    speed           TEXT,
    is_uplink       BOOLEAN DEFAULT FALSE,
    mac_restriction MACADDR,
    notes           TEXT,
    UNIQUE(switch_id, port_number)
);

CREATE TABLE IF NOT EXISTS patch_panels (
    id              BIGSERIAL PRIMARY KEY,
    site_id         BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    total_ports     INTEGER NOT NULL DEFAULT 24,
    location_id     BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    notes           TEXT,
    UNIQUE(site_id, name)
);

CREATE TABLE IF NOT EXISTS patch_panel_ports (
    id                  BIGSERIAL PRIMARY KEY,
    patch_panel_id      BIGINT NOT NULL REFERENCES patch_panels(id) ON DELETE CASCADE,
    port_number         INTEGER NOT NULL,
    port_label          TEXT,
    linked_port_id      BIGINT REFERENCES patch_panel_ports(id) ON DELETE SET NULL,
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
    hostname            TEXT NOT NULL,
    dns_name            TEXT,
    serial_number       TEXT,
    asset_tag           TEXT,
    category_id         BIGINT NOT NULL REFERENCES categories(id),
    status              TEXT NOT NULL DEFAULT 'planned',
    is_up               BOOLEAN DEFAULT TRUE,
    os_id               BIGINT REFERENCES operating_systems(id) ON DELETE SET NULL,
    has_rmm             BOOLEAN DEFAULT FALSE,
    has_antivirus       BOOLEAN DEFAULT FALSE,
    supplier_id         BIGINT REFERENCES suppliers(id) ON DELETE SET NULL,
    installation_date   DATE,
    is_reserved         BOOLEAN DEFAULT FALSE,
    notes               TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, hostname)
);

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
-- NETWORK INTERFACES
-- ============================================================

CREATE TABLE IF NOT EXISTS device_interfaces (
    id          BIGSERIAL PRIMARY KEY,
    device_id   BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    mac_address MACADDR,
    notes       TEXT,
    UNIQUE(device_id, name)
);

-- ============================================================
-- IP ADDRESSES
-- ============================================================

CREATE TABLE IF NOT EXISTS device_ips (
    id              BIGSERIAL PRIMARY KEY,
    interface_id    BIGINT NOT NULL REFERENCES device_interfaces(id) ON DELETE CASCADE,
    ip_address      INET NOT NULL,
    vlan_id         BIGINT REFERENCES vlans(id) ON DELETE SET NULL,
    is_primary      BOOLEAN DEFAULT FALSE,
    notes           TEXT
);

-- Deferred FK: vlans.gateway_device_ip_id -> device_ips (circular dep).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_vlans_gateway_device_ip' AND table_name = 'vlans'
    ) THEN
        ALTER TABLE vlans
            ADD CONSTRAINT fk_vlans_gateway_device_ip
            FOREIGN KEY (gateway_device_ip_id) REFERENCES device_ips(id) ON DELETE SET NULL;
    END IF;
END
$$;

-- ============================================================
-- PHYSICAL CONNECTIONS
-- ============================================================

CREATE TABLE IF NOT EXISTS device_connections (
    id                  BIGSERIAL PRIMARY KEY,
    interface_id        BIGINT NOT NULL REFERENCES device_interfaces(id) ON DELETE CASCADE,
    switch_port_id      BIGINT REFERENCES switch_ports(id) ON DELETE SET NULL,
    patch_panel_port_id BIGINT REFERENCES patch_panel_ports(id) ON DELETE SET NULL,
    connected_at        DATE,
    notes               TEXT
);

-- ============================================================
-- AUDIT LOG
-- ============================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    username    TEXT NOT NULL,
    action      TEXT NOT NULL,
    resource    TEXT NOT NULL,
    resource_id BIGINT,
    detail      TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user     ON audit_logs(user_id);

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_devices_site             ON devices(site_id);
CREATE INDEX IF NOT EXISTS idx_devices_hostname         ON devices(hostname);
CREATE INDEX IF NOT EXISTS idx_device_interfaces_device ON device_interfaces(device_id);
CREATE INDEX IF NOT EXISTS idx_device_ips_address       ON device_ips(ip_address);
CREATE UNIQUE INDEX IF NOT EXISTS idx_switch_ports_mac    ON switch_ports(mac_restriction) WHERE mac_restriction IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_switch_ports_sw          ON switch_ports(switch_id);
CREATE INDEX IF NOT EXISTS idx_address_blocks_site      ON address_blocks(site_id);
CREATE INDEX IF NOT EXISTS idx_vlans_block              ON vlans(address_block_id);
