BEGIN;

-- Allow patch panel ports to link directly to a switch port
ALTER TABLE patch_panel_ports
    ADD COLUMN switch_port_id BIGINT REFERENCES switch_ports(id) ON DELETE SET NULL;

-- 1:1 enforcement: each switch port can be linked to at most one patch panel port
CREATE UNIQUE INDEX uq_pp_switch_port
    ON patch_panel_ports (switch_port_id) WHERE switch_port_id IS NOT NULL;

-- 1:1 enforcement: each switch port can appear in at most one device connection
CREATE UNIQUE INDEX uq_dc_switch_port
    ON device_connections (switch_port_id) WHERE switch_port_id IS NOT NULL;

-- 1:1 enforcement: each patch panel port can appear in at most one device connection
CREATE UNIQUE INDEX uq_dc_patch_panel_port
    ON device_connections (patch_panel_port_id) WHERE patch_panel_port_id IS NOT NULL;

-- 1:1 enforcement: each interface can have at most one connection
CREATE UNIQUE INDEX uq_dc_interface
    ON device_connections (interface_id);

COMMIT;
