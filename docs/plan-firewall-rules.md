# Firewall Rules Implementation Plan

## Context
For disaster recovery documentation — if a site burns down, we need to rebuild the firewall config. VLANs are already tracked; now we need to track firewall rules (who can talk to whom, on which ports/protocols).

## Branch: `feat/firewall-rules` (from main, after device-groups merge)

## Migration `004_firewall_rules.up.sql`
```sql
CREATE TABLE firewall_rules (
    id            BIGSERIAL PRIMARY KEY,
    site_id       BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

    -- Source endpoint (at most one set; all NULL = "any")
    src_device_id BIGINT REFERENCES devices(id) ON DELETE CASCADE,
    src_group_id  BIGINT REFERENCES device_groups(id) ON DELETE CASCADE,
    src_vlan_id   BIGINT REFERENCES vlans(id) ON DELETE CASCADE,
    src_cidr      CIDR,  -- free-form IP/subnet for external addresses

    -- Destination endpoint (same logic)
    dst_device_id BIGINT REFERENCES devices(id) ON DELETE CASCADE,
    dst_group_id  BIGINT REFERENCES device_groups(id) ON DELETE CASCADE,
    dst_vlan_id   BIGINT REFERENCES vlans(id) ON DELETE CASCADE,
    dst_cidr      CIDR,

    -- Ports (separate src/dst)
    src_port      TEXT NOT NULL DEFAULT '*',
    dst_port      TEXT NOT NULL DEFAULT '*',

    protocol      TEXT NOT NULL DEFAULT 'any',   -- tcp, udp, both, icmp, any
    action        TEXT NOT NULL DEFAULT 'allow',  -- allow, deny
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
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/firewall-rules?site_id=` | List rules (ordered by position) |
| POST | `/firewall-rules` | Create rule |
| GET | `/firewall-rules/:id` | Get rule |
| PUT | `/firewall-rules/:id` | Update rule |
| DELETE | `/firewall-rules/:id` | Delete rule |
| GET | `/sites/:id/firewall-rules` | Rules by site |

## Validation
- At most one src endpoint set (device/group/vlan/cidr), same for dst; all NULL = "any"
- All FK-referenced entities must belong to the rule's site
- `src_cidr`/`dst_cidr` validated as valid CIDR by PostgreSQL's CIDR type
- protocol in (tcp, udp, both, icmp, any); action in (allow, deny)
- Position: auto-assign `MAX(position)+1` on create if not provided

## Files to Create
- `backend/db/migrations/004_firewall_rules.{up,down}.sql`
- `backend/models/firewall_rule.go`
- `backend/handlers/firewall_rules.go`
- `cli/internal/tui/resource/firewall_rules.go`

## Files to Modify
- `backend/router.go` — register handler + routes
- `cli/internal/tui/menu.go` — add menu entry
- Docs: `CLAUDE.md`, `docs/API.md`

## CLI Resource
- Source/destination fields use pickers for device/group/vlan (filtered by site)
- CIDR is a free-text field
- PickerOptions for protocol, action, enabled
- Table shows: position, action, src summary, dst summary, ports, protocol, enabled

## Versioning
- Merge to main → tag `v0.6.0`
