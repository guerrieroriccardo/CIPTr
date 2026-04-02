-- Migration 013: Consolidate switches and patch panels into the devices table.
--
-- After this migration:
--   - switches and patch_panels tables are dropped
--   - switch_ports.switch_id -> switch_ports.device_id (FK to devices)
--   - patch_panel_ports.patch_panel_id -> patch_panel_ports.device_id (FK to devices)
--   - categories gains a port_type column ('switch', 'patch_panel', or NULL)
--   - devices gains a total_ports column (nullable)

BEGIN;

-- 1. Add new columns
ALTER TABLE devices ADD COLUMN IF NOT EXISTS total_ports INTEGER;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS port_type TEXT;

-- 2. Set port_type for known categories (case-insensitive match)
UPDATE categories SET port_type = 'switch' WHERE LOWER(name) = 'switch';
UPDATE categories SET port_type = 'patch_panel' WHERE LOWER(name) = 'patch panel';

-- 3. Ensure a "Switch" category exists (in case it doesn't)
INSERT INTO categories (name, short_code, port_type)
SELECT 'Switch', 'SW', 'switch'
WHERE NOT EXISTS (SELECT 1 FROM categories WHERE LOWER(name) = 'switch');

-- 4. Ensure a "Patch Panel" category exists
INSERT INTO categories (name, short_code, port_type)
SELECT 'Patch Panel', 'PP', 'patch_panel'
WHERE NOT EXISTS (SELECT 1 FROM categories WHERE LOWER(name) = 'patch panel');

-- 5. Migrate switches -> devices
-- Use a temp table to track old_id -> new_id mapping.
CREATE TEMP TABLE _switch_id_map (old_id BIGINT PRIMARY KEY, new_id BIGINT NOT NULL);

WITH inserted AS (
    INSERT INTO devices (site_id, location_id, model_id, hostname, category_id, status, total_ports, notes)
    SELECT s.site_id, s.location_id, s.model_id, s.hostname,
           (SELECT id FROM categories WHERE port_type = 'switch' LIMIT 1),
           'active', s.total_ports, s.notes
    FROM switches s
    RETURNING id, hostname, site_id
)
INSERT INTO _switch_id_map (old_id, new_id)
SELECT s.id, i.id
FROM switches s
JOIN inserted i ON i.hostname = s.hostname AND i.site_id = s.site_id;

-- 5b. Create device_interface + device_ip for switches that had an IP address
INSERT INTO device_interfaces (device_id, name)
SELECT m.new_id, 'mgmt'
FROM _switch_id_map m
JOIN switches s ON s.id = m.old_id
WHERE s.ip_address IS NOT NULL;

INSERT INTO device_ips (interface_id, ip_address, vlan_id, is_primary)
SELECT di.id, s.ip_address, s.vlan_id, TRUE
FROM _switch_id_map m
JOIN switches s ON s.id = m.old_id
JOIN device_interfaces di ON di.device_id = m.new_id AND di.name = 'mgmt'
WHERE s.ip_address IS NOT NULL;

-- 6. Migrate patch_panels -> devices
CREATE TEMP TABLE _pp_id_map (old_id BIGINT PRIMARY KEY, new_id BIGINT NOT NULL);

WITH inserted AS (
    INSERT INTO devices (site_id, location_id, hostname, category_id, status, total_ports, notes)
    SELECT pp.site_id, pp.location_id, pp.name,
           (SELECT id FROM categories WHERE port_type = 'patch_panel' LIMIT 1),
           'active', pp.total_ports, pp.notes
    FROM patch_panels pp
    RETURNING id, hostname, site_id
)
INSERT INTO _pp_id_map (old_id, new_id)
SELECT pp.id, i.id
FROM patch_panels pp
JOIN inserted i ON i.hostname = pp.name AND i.site_id = pp.site_id;

-- 7. Remap switch_ports: rename column and update FK
-- First drop the old FK and unique constraints
ALTER TABLE switch_ports DROP CONSTRAINT IF EXISTS switch_ports_switch_id_fkey;
ALTER TABLE switch_ports DROP CONSTRAINT IF EXISTS switch_ports_switch_id_port_number_key;
DROP INDEX IF EXISTS idx_switch_ports_sw;

-- Rename the column
ALTER TABLE switch_ports RENAME COLUMN switch_id TO device_id;

-- Update the IDs to point to the new device IDs
UPDATE switch_ports sp
SET device_id = m.new_id
FROM _switch_id_map m
WHERE sp.device_id = m.old_id;

-- Re-add constraints pointing to devices
ALTER TABLE switch_ports ADD CONSTRAINT switch_ports_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE;
ALTER TABLE switch_ports ADD CONSTRAINT switch_ports_device_id_port_number_key
    UNIQUE (device_id, port_number);
CREATE INDEX idx_switch_ports_dev ON switch_ports(device_id);

-- 8. Remap patch_panel_ports: rename column and update FK
ALTER TABLE patch_panel_ports DROP CONSTRAINT IF EXISTS patch_panel_ports_patch_panel_id_fkey;
ALTER TABLE patch_panel_ports DROP CONSTRAINT IF EXISTS patch_panel_ports_patch_panel_id_port_number_key;

ALTER TABLE patch_panel_ports RENAME COLUMN patch_panel_id TO device_id;

UPDATE patch_panel_ports ppp
SET device_id = m.new_id
FROM _pp_id_map m
WHERE ppp.device_id = m.old_id;

ALTER TABLE patch_panel_ports ADD CONSTRAINT patch_panel_ports_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE;
ALTER TABLE patch_panel_ports ADD CONSTRAINT patch_panel_ports_device_id_port_number_key
    UNIQUE (device_id, port_number);

-- 9. Drop old tables (cascades handle remaining FKs)
DROP TABLE IF EXISTS switches CASCADE;
DROP TABLE IF EXISTS patch_panels CASCADE;

-- 10. Clean up temp tables
DROP TABLE IF EXISTS _switch_id_map;
DROP TABLE IF EXISTS _pp_id_map;

COMMIT;
