# CIPTr Backend API Reference

> For frontend implementation. Base URL: `http://localhost:8080/api/v1`

## Response Envelope

```json
// Success
{ "data": <payload>, "error": null }

// Error
{ "data": null, "error": "message" }
```

- Creates return `201`, everything else `200`
- Delete returns `{ "data": {"deleted": true} }`
- Not found returns `404`
- Validation errors return `400`
- All `PUT` endpoints do full replacement (not partial)

---

## Resources

### Clients

| Method | Path | Notes |
|--------|------|-------|
| GET | `/clients` | ordered by name |
| POST | `/clients` | |
| GET | `/clients/:id` | |
| PUT | `/clients/:id` | |
| DELETE | `/clients/:id` | cascades to sites → all descendants |
| GET | `/clients/:id/sites` | |

**Required:** `name`, `short_code`
**Unique:** `name` (global), `short_code` (global)

```json
{ "name": "Acme Corp", "short_code": "ACM", "notes": null }
```

**Frontend note:** Short code should auto-generate from name (3 uppercase chars):
- 3-char name → use as-is
- 1 word → first 3 consonants (BERPA → BRP)
- 2 words → first 2 consonants of word 1 + first char of word 2 (BRAND-EAT → BRE)
- 3+ words → first char of first 3 words (Officine Meccaniche Pontina → OMP)

---

### Sites

| Method | Path | Notes |
|--------|------|-------|
| GET | `/sites` | `?client_id=` filter |
| POST | `/sites` | |
| GET | `/sites/:id` | |
| PUT | `/sites/:id` | |
| DELETE | `/sites/:id` | cascades to locations, vlans, devices, switches, panels |
| GET | `/sites/:id/address-blocks` | |
| GET | `/sites/:id/vlans` | |
| GET | `/sites/:id/locations` | |
| GET | `/sites/:id/devices` | |
| GET | `/sites/:id/switches` | |
| GET | `/sites/:id/patch-panels` | |

**Required:** `client_id`, `name`
**Unique:** `(client_id, name)`

```json
{ "client_id": 1, "name": "Main Office", "address": "Via Roma 1", "notes": null }
```

---

### Locations

| Method | Path | Notes |
|--------|------|-------|
| GET | `/locations` | `?site_id=` filter |
| POST/PUT/DELETE | `/locations/:id` | |

**Required:** `site_id`, `name`
**Unique:** `(site_id, name)`

```json
{ "site_id": 1, "name": "Server Room", "floor": "1", "notes": null }
```

---

### Address Blocks

| Method | Path | Notes |
|--------|------|-------|
| GET | `/address-blocks` | `?site_id=` filter |
| POST/PUT/DELETE | `/address-blocks/:id` | |
| GET | `/address-blocks/:id/vlans` | |

**Required:** `site_id`, `network` (CIDR notation)
**Unique:** `(site_id, network)`

**Validation:**
- `network` must not overlap with existing blocks in the same site (uses PostgreSQL CIDR `&&` operator)

```json
{ "site_id": 1, "network": "10.10.0.0/20", "description": "Main block", "notes": null }
```

---

### VLANs

| Method | Path | Notes |
|--------|------|-------|
| GET | `/vlans` | `?site_id=` and/or `?address_block_id=` filters |
| POST/PUT/DELETE | `/vlans/:id` | |

**Required:** `site_id`, `vlan_id` (tag number), `name`
**Unique:** `(site_id, vlan_id)`

**Validation (all return 400):**
1. VLAN tag must be unique within the site
2. Subnet must be valid CIDR
3. Subnet must not overlap with other VLANs in the same site
4. Subnet must fit entirely within address block (if both provided)

```json
{
  "site_id": 1, "address_block_id": 1, "vlan_id": 10,
  "name": "Users", "subnet": "10.10.0.0/24",
  "gateway_device_ip_id": 7, "description": null
}
```

---

### Manufacturers

Standard CRUD at `/manufacturers`. **Required:** `name` (globally unique).

### Categories

Standard CRUD at `/categories`. **Required:** `name` (globally unique). Optional: `short_code`.

### Suppliers

Standard CRUD at `/suppliers`. **Required:** `name`.

```json
{ "name": "TechSupply", "address": "Via Roma 1", "phone": "+39...", "email": "info@..." }
```

---

### Device Models

| Method | Path | Notes |
|--------|------|-------|
| GET | `/device-models` | `?category_id=` and/or `?manufacturer_id=` filters |
| POST/PUT/DELETE | `/device-models/:id` | |

**Required:** `manufacturer_id`, `model_name`, `category_id`
**Unique:** `(manufacturer_id, model_name)`

```json
{
  "manufacturer_id": 1, "model_name": "ProLiant DL380", "category_id": 1,
  "os_default": "Windows Server 2022", "specs": "32GB RAM", "notes": null
}
```

**Frontend note:** Display as "Manufacturer ModelName" (e.g. "HP ProLiant DL380").

---

### Devices

| Method | Path | Notes |
|--------|------|-------|
| GET | `/devices` | `?site_id=`, `?status=`, `?category_id=`, `?search=` (all combinable). `search` matches hostname or dns_name (case-insensitive) |
| POST/PUT/DELETE | `/devices/:id` | |
| GET | `/devices/:id/interfaces` | |
| GET | `/devices/:id/ips` | IPs across all interfaces |
| GET | `/devices/:id/connections` | connections across all interfaces |

**Required:** `site_id`, `hostname`, `category_id`
**Unique:** `(site_id, hostname)`
**Default:** `status` = `"active"` if omitted

**Validation:**
- Hostname must be unique within the site

**Status values:** `active`, `planned`, `inactive`, `decommissioned`, `storage`

```json
{
  "site_id": 1, "hostname": "SRV-DC01", "category_id": 1,
  "location_id": 1, "model_id": 1,
  "dns_name": "srv-dc01.domain.local", "serial_number": "ABC123",
  "asset_tag": "IT-001", "status": "active", "is_up": true,
  "os": "Windows Server 2022", "has_rmm": true, "has_antivirus": true,
  "supplier_id": 1, "installation_date": "2024-01-15",
  "is_reserved": false, "notes": null
}
```

---

### Device Interfaces (NICs)

| Method | Path | Notes |
|--------|------|-------|
| GET | `/device-interfaces` | `?device_id=` filter |
| POST/PUT/DELETE | `/device-interfaces/:id` | |

**Required:** `device_id`, `name`
**Unique:** `(device_id, name)`

```json
{ "device_id": 1, "name": "eth0", "mac_address": "00:11:22:33:44:55", "notes": null }
```

---

### Device IPs

| Method | Path | Notes |
|--------|------|-------|
| GET | `/device-ips` | `?interface_id=` and/or `?vlan_id=` filters |
| POST/PUT/DELETE | `/device-ips/:id` | |

**Required:** `interface_id`, `ip_address`

**Validation:**
- If `vlan_id` is set and the VLAN has a subnet, the IP must be within that subnet

```json
{ "interface_id": 1, "ip_address": "10.10.0.10", "vlan_id": 1, "is_primary": true, "notes": null }
```

---

### Device Connections

| Method | Path | Notes |
|--------|------|-------|
| GET | `/device-connections` | `?interface_id=`, `?switch_port_id=`, `?patch_panel_port_id=` filters |
| POST/PUT/DELETE | `/device-connections/:id` | |

**Required:** `interface_id`

```json
{
  "interface_id": 1, "switch_port_id": 1, "patch_panel_port_id": 1,
  "connected_at": "2024-01-15", "notes": null
}
```

---

### Switches

| Method | Path | Notes |
|--------|------|-------|
| GET | `/switches` | `?site_id=` filter |
| POST/PUT/DELETE | `/switches/:id` | |
| GET | `/switches/:id/ports` | |

**Required:** `site_id`, `name`
**Unique:** `(site_id, name)`
**Default:** `total_ports` = `24` if omitted
**Auto:** Creating a switch auto-creates `total_ports` port rows (Port 1 … Port N)

```json
{
  "site_id": 1, "name": "SW-CORE-01", "model_id": 1,
  "ip_address": "10.10.0.254", "location": "Rack A",
  "total_ports": 48, "notes": null
}
```

---

### Switch Ports

| Method | Path | Notes |
|--------|------|-------|
| GET | `/switch-ports` | `?switch_id=` filter |
| POST/PUT/DELETE | `/switch-ports/:id` | |

**Required:** `switch_id`, `port_number`
**Unique:** `(switch_id, port_number)`

```json
{ "switch_id": 1, "port_number": 1, "port_label": "Gi0/1", "speed": "1G", "is_uplink": false, "notes": null }
```

---

### Patch Panels

| Method | Path | Notes |
|--------|------|-------|
| GET | `/patch-panels` | `?site_id=` filter |
| POST/PUT/DELETE | `/patch-panels/:id` | |
| GET | `/patch-panels/:id/ports` | |

**Required:** `site_id`, `name`
**Unique:** `(site_id, name)`
**Default:** `total_ports` = `24` if omitted
**Auto:** Creating a patch panel auto-creates `total_ports` port rows (Port 1 … Port N)

---

### Patch Panel Ports

| Method | Path | Notes |
|--------|------|-------|
| GET | `/patch-panel-ports` | `?patch_panel_id=` filter |
| POST/PUT/DELETE | `/patch-panel-ports/:id` | |

**Required:** `patch_panel_id`, `port_number`
**Unique:** `(patch_panel_id, port_number)`

---

## Cascade Summary

| Deleted | Children Effect |
|---------|----------------|
| Client | Sites → CASCADE |
| Site | Locations, address blocks, VLANs, devices, switches, patch panels → CASCADE |
| Location | Devices: `location_id` → NULL |
| Address block | VLANs: `address_block_id` → NULL |
| Device | Interfaces → CASCADE (→ IPs, connections CASCADE) |
| Device model | Devices, switches: `model_id` → NULL |
| Switch | Ports → CASCADE |
| Patch panel | Ports → CASCADE |
| Supplier | Devices: `supplier_id` → NULL |

---

## FK Picker Fields (for frontend dropdowns)

These fields should use select/dropdown components instead of free text:

| Resource | Field | Options source |
|----------|-------|---------------|
| Site | `client_id` | GET /clients |
| Location | `site_id` | GET /sites |
| Address block | `site_id` | GET /sites |
| VLAN | `site_id`, `address_block_id`, `gateway_device_ip_id` | GET /sites, GET /address-blocks, GET /device-ips |
| Device model | `manufacturer_id`, `category_id` | GET /manufacturers, GET /categories |
| Device | `site_id`, `category_id`, `location_id`, `model_id`, `supplier_id` | respective endpoints |
| Device | `status` | static: active, planned, inactive, decommissioned, storage |
| Device | `is_up`, `has_rmm`, `has_antivirus`, `is_reserved` | static: true, false |
| Device interface | `device_id` | GET /devices |
| Device IP | `interface_id`, `vlan_id` | GET /device-interfaces, GET /vlans |
| Device connection | `interface_id`, `switch_port_id`, `patch_panel_port_id` | respective endpoints |
| Switch | `site_id`, `model_id` | GET /sites, GET /device-models |
| Switch port | `switch_id` | GET /switches |
| Patch panel | `site_id` | GET /sites |
| Patch panel port | `patch_panel_id` | GET /patch-panels |

**Display hints:**
- Device models: show as "Manufacturer ModelName" (e.g. "HP ProLiant DL380")
- Switch/patch panel ports: show `port_label` if set, otherwise "Port N"
- Address blocks: show `network` (CIDR)
- VLANs: show `name`

## Health

`GET /api/v1/health` → `{"status": "ok"}` (not wrapped in envelope)
