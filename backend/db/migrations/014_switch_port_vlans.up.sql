-- Migration 014: Add VLAN tagging and disabled flag to switch ports.
--
-- Each switch port can have:
--   - One untagged (native) VLAN
--   - Multiple tagged (trunked) VLANs
--   - A disabled flag (administratively shut down)

ALTER TABLE switch_ports ADD COLUMN IF NOT EXISTS untagged_vlan_id BIGINT
    REFERENCES vlans(id) ON DELETE SET NULL;

ALTER TABLE switch_ports ADD COLUMN IF NOT EXISTS is_disabled BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS switch_port_tagged_vlans (
    switch_port_id BIGINT NOT NULL REFERENCES switch_ports(id) ON DELETE CASCADE,
    vlan_id        BIGINT NOT NULL REFERENCES vlans(id) ON DELETE CASCADE,
    PRIMARY KEY (switch_port_id, vlan_id)
);
