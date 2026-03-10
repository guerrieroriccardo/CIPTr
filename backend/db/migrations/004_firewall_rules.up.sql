CREATE TABLE firewall_rules (
    id            BIGSERIAL PRIMARY KEY,
    site_id       BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

    -- Source endpoint (at most one set; all NULL = "any")
    src_device_id BIGINT REFERENCES devices(id) ON DELETE CASCADE,
    src_group_id  BIGINT REFERENCES device_groups(id) ON DELETE CASCADE,
    src_vlan_id   BIGINT REFERENCES vlans(id) ON DELETE CASCADE,
    src_cidr      CIDR,

    -- Destination endpoint (same logic)
    dst_device_id BIGINT REFERENCES devices(id) ON DELETE CASCADE,
    dst_group_id  BIGINT REFERENCES device_groups(id) ON DELETE CASCADE,
    dst_vlan_id   BIGINT REFERENCES vlans(id) ON DELETE CASCADE,
    dst_cidr      CIDR,

    -- Ports (separate src/dst)
    src_port      TEXT NOT NULL DEFAULT '*',
    dst_port      TEXT NOT NULL DEFAULT '*',

    protocol      TEXT NOT NULL DEFAULT 'any',
    action        TEXT NOT NULL DEFAULT 'allow',
    position      INTEGER NOT NULL DEFAULT 0,
    enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    description   TEXT,
    notes         TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),

    -- At most one src endpoint
    CONSTRAINT chk_src_single CHECK (
        (CASE WHEN src_device_id IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN src_group_id  IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN src_vlan_id   IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN src_cidr      IS NOT NULL THEN 1 ELSE 0 END) <= 1
    ),
    -- At most one dst endpoint
    CONSTRAINT chk_dst_single CHECK (
        (CASE WHEN dst_device_id IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN dst_group_id  IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN dst_vlan_id   IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN dst_cidr      IS NOT NULL THEN 1 ELSE 0 END) <= 1
    )
);
