#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080/api/v1}"

echo "==> Seeding via API at $API_URL"

# Check API is reachable
if ! curl -sf "$API_URL/health" >/dev/null 2>&1; then
  echo "ERROR: API not reachable at $API_URL"
  exit 1
fi

post() {
  local path="$1"
  local data="$2"
  local result
  result=$(curl -sf -X POST "$API_URL$path" -H 'Content-Type: application/json' --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzI5NzkyNjEsInVzZXJfaWQiOjIsInVzZXJuYW1lIjoiZ3VlcnJvIn0.ZrCGayGgM-a4Nu72p9fUocgZOl04h9T0EC-W9hwT1fQ" -d "$data" 2>&1) || {
    echo "  FAILED: POST $path"
    echo "  Data: $data"
    echo "  Response: $result"
    return 1
  }
  echo "$result"
}

# ── Manufacturers ──
echo "==> Manufacturers..."
post /manufacturers '{"name":"HP"}'
post /manufacturers '{"name":"Cisco"}'
post /manufacturers '{"name":"Dell"}'
post /manufacturers '{"name":"Ubiquiti"}'
post /manufacturers '{"name":"APC"}'
post /manufacturers '{"name":"Lenovo"}'
post /manufacturers '{"name":"Fortinet"}'

# ── Categories ──
echo "==> Categories..."
post /categories '{"name":"Server","short_code":"SRV"}'
post /categories '{"name":"PC Desktop","short_code":"PC"}'
post /categories '{"name":"Notebook","short_code":"NB"}'
post /categories '{"name":"Switch","short_code":"SW"}'
post /categories '{"name":"Access Point","short_code":"AP"}'
post /categories '{"name":"Firewall","short_code":"FW"}'
post /categories '{"name":"UPS","short_code":"UPS"}'
post /categories '{"name":"Printer","short_code":"PRT"}'
post /categories '{"name":"NAS","short_code":"NAS"}'
post /categories '{"name":"Router","short_code":"RTR"}'

# ── Suppliers ──
echo "==> Suppliers..."
post /suppliers '{"name":"TechDistribution Srl","address":"Via dell'\''Industria 5, Bologna","phone":"+39 051 1234567","email":"ordini@techdist.it"}'
post /suppliers '{"name":"InfoStore SpA","address":"Via Giardini 300, Modena","phone":"+39 059 9876543","email":"vendite@infostore.it"}'

# ── Clients ──
echo "==> Clients..."
post /clients '{"name":"Berpa Costruzioni","short_code":"BRP","domain":"berpa.local","notes":"Construction company"}'
post /clients '{"name":"Officine Meccaniche Pontina","short_code":"OMP","domain":"omp.local","notes":"Mechanical workshop"}'
post /clients '{"name":"Studio Legale Rossi","short_code":"SLR","notes":"Law firm"}'
post /clients '{"name":"Farmacia Centrale","short_code":"FRC","notes":"Pharmacy chain"}'

# ── Sites ──
echo "==> Sites..."
post /sites '{"client_id":1,"name":"Sede Principale","address":"Via Roma 15, Sassuolo (MO)","notes":"HQ"}'
post /sites '{"client_id":1,"name":"Cantiere Nord","address":"Via Emilia 200, Modena","notes":"Construction site office"}'
post /sites '{"client_id":2,"name":"Stabilimento","address":"Via Pontina km 42, Latina"}'
post /sites '{"client_id":3,"name":"Ufficio Centro","address":"Corso Italia 8, Sassuolo (MO)"}'
post /sites '{"client_id":4,"name":"Farmacia 1","address":"Piazza Garibaldi 3, Sassuolo (MO)"}'
post /sites '{"client_id":4,"name":"Farmacia 2","address":"Via Radici 120, Sassuolo (MO)"}'

# ── Locations ──
echo "==> Locations..."
post /locations '{"site_id":1,"name":"Sala Server","floor":"0","notes":"Basement server room"}'
post /locations '{"site_id":1,"name":"Ufficio Direzione","floor":"1"}'
post /locations '{"site_id":1,"name":"Open Space","floor":"1"}'
post /locations '{"site_id":3,"name":"Sala CED","floor":"0","notes":"Small rack room"}'
post /locations '{"site_id":4,"name":"Ufficio","floor":"0"}'
post /locations '{"site_id":5,"name":"Retro Banco","floor":"0","notes":"Behind the counter"}'

# ── Device Models ──
echo "==> Device Models..."
post /device-models '{"manufacturer_id":1,"model_name":"ProLiant DL380 Gen10","category_id":1,"os_default":"Windows Server 2022","specs":"2x Xeon Gold, 64GB RAM, 4x 1.2TB SAS"}'
post /device-models '{"manufacturer_id":1,"model_name":"ProLiant DL360 Gen10","category_id":1,"os_default":"Windows Server 2022","specs":"Xeon Silver, 32GB RAM, 2x 480GB SSD"}'
post /device-models '{"manufacturer_id":3,"model_name":"OptiPlex 7090","category_id":2,"os_default":"Windows 11 Pro","specs":"i7-11700, 16GB RAM, 512GB NVMe"}'
post /device-models '{"manufacturer_id":6,"model_name":"ThinkPad T14s Gen3","category_id":3,"os_default":"Windows 11 Pro","specs":"i7-1260P, 16GB RAM, 512GB NVMe"}'
post /device-models '{"manufacturer_id":2,"model_name":"Catalyst 2960-X 48","category_id":4,"specs":"48x 1GbE, 4x SFP+"}'
post /device-models '{"manufacturer_id":2,"model_name":"Catalyst 2960-X 24","category_id":4,"specs":"24x 1GbE, 4x SFP+"}'
post /device-models '{"manufacturer_id":4,"model_name":"UniFi U6 Pro","category_id":5,"specs":"Wi-Fi 6, PoE"}'
post /device-models '{"manufacturer_id":7,"model_name":"FortiGate 60F","category_id":6,"os_default":"FortiOS 7.4","specs":"10 GbE ports, 700 Mbps throughput"}'
post /device-models '{"manufacturer_id":5,"model_name":"Smart-UPS 1500","category_id":7,"specs":"1500VA / 1000W, LCD, rack 2U"}'
post /device-models '{"manufacturer_id":1,"model_name":"LaserJet Pro M404dn","category_id":8,"specs":"B&W, duplex, network"}'
post /device-models '{"manufacturer_id":3,"model_name":"PowerEdge R640","category_id":1,"os_default":"Proxmox VE 8","specs":"2x Xeon Gold, 128GB RAM, 8x 960GB SSD","notes":"Virtualization host"}'

# ── Address Blocks ──
echo "==> Address Blocks..."
post /address-blocks '{"site_id":1,"network":"10.10.0.0/20","description":"Berpa HQ main block"}'
post /address-blocks '{"site_id":3,"network":"10.20.0.0/22","description":"OMP factory block"}'
post /address-blocks '{"site_id":4,"network":"192.168.1.0/24","description":"Studio Rossi single subnet"}'
post /address-blocks '{"site_id":5,"network":"192.168.10.0/24","description":"Farmacia 1"}'
post /address-blocks '{"site_id":6,"network":"192.168.11.0/24","description":"Farmacia 2"}'

# ── VLANs ──
echo "==> VLANs..."
# Berpa HQ
post /vlans '{"site_id":1,"address_block_id":1,"vlan_id":1,"name":"Management","subnet":"10.10.0.0/24","description":"Network devices management"}'
post /vlans '{"site_id":1,"address_block_id":1,"vlan_id":10,"name":"Servers","subnet":"10.10.1.0/24","description":"Server VLAN"}'
post /vlans '{"site_id":1,"address_block_id":1,"vlan_id":20,"name":"Users","subnet":"10.10.2.0/24","description":"Workstations"}'
post /vlans '{"site_id":1,"address_block_id":1,"vlan_id":30,"name":"VoIP","subnet":"10.10.3.0/24","description":"IP phones"}'
post /vlans '{"site_id":1,"address_block_id":1,"vlan_id":40,"name":"Guest WiFi","subnet":"10.10.4.0/24","description":"Guest wireless network"}'
post /vlans '{"site_id":1,"address_block_id":1,"vlan_id":99,"name":"Printers","subnet":"10.10.9.0/24","description":"Printers and MFPs"}'
# OMP factory
post /vlans '{"site_id":3,"address_block_id":2,"vlan_id":1,"name":"Management","subnet":"10.20.0.0/24"}'
post /vlans '{"site_id":3,"address_block_id":2,"vlan_id":10,"name":"Servers","subnet":"10.20.1.0/24"}'
post /vlans '{"site_id":3,"address_block_id":2,"vlan_id":20,"name":"Office","subnet":"10.20.2.0/24","description":"Office workstations"}'
post /vlans '{"site_id":3,"address_block_id":2,"vlan_id":30,"name":"Production","subnet":"10.20.3.0/24","description":"Factory floor devices"}'
# Studio Rossi
post /vlans '{"site_id":4,"address_block_id":3,"vlan_id":1,"name":"LAN","subnet":"192.168.1.0/24","description":"Single flat network"}'
# Farmacie
post /vlans '{"site_id":5,"address_block_id":4,"vlan_id":1,"name":"LAN","subnet":"192.168.10.0/24"}'
post /vlans '{"site_id":6,"address_block_id":5,"vlan_id":1,"name":"LAN","subnet":"192.168.11.0/24"}'

# ── Switches (ports auto-created by API) ──
echo "==> Switches..."
post /switches '{"site_id":1,"hostname":"SW001","model_id":5,"ip_address":"10.10.0.10","location_id":1,"total_ports":48,"notes":"Core switch"}'
post /switches '{"site_id":1,"hostname":"SW002","model_id":6,"ip_address":"10.10.0.11","location_id":1,"total_ports":24,"notes":"Floor 1 access switch"}'
post /switches '{"site_id":3,"hostname":"SW001","model_id":6,"ip_address":"10.20.0.10","location_id":4,"total_ports":24}'
post /switches '{"site_id":4,"hostname":"SW001","model_id":6,"ip_address":"192.168.1.2","location_id":5,"total_ports":24}'

# ── Patch Panels (ports auto-created by API) ──
echo "==> Patch Panels..."
post /patch-panels '{"site_id":1,"name":"PP-RACK-A-1","total_ports":24,"location_id":1,"notes":"Top of rack"}'
post /patch-panels '{"site_id":1,"name":"PP-RACK-A-2","total_ports":24,"location_id":1}'

# ── Devices ──
echo "==> Devices..."
# Berpa HQ servers
post /devices '{"site_id":1,"location_id":1,"model_id":1,"hostname":"SRV-DC01","dns_name":"srv-dc01.berpa.local","serial_number":"CZJ12345AB","asset_tag":"IT-001","category_id":1,"status":"active","is_up":true,"os":"Windows Server 2022","has_rmm":true,"has_antivirus":true,"supplier_id":1,"installation_date":"2023-06-15","notes":"Primary domain controller"}'
post /devices '{"site_id":1,"location_id":1,"model_id":2,"hostname":"SRV-DC02","dns_name":"srv-dc02.berpa.local","serial_number":"CZJ12345AC","asset_tag":"IT-002","category_id":1,"status":"active","is_up":true,"os":"Windows Server 2022","has_rmm":true,"has_antivirus":true,"supplier_id":1,"installation_date":"2023-06-15","notes":"Secondary domain controller"}'
post /devices '{"site_id":1,"location_id":1,"model_id":11,"hostname":"SRV-PROX01","dns_name":"srv-prox01.berpa.local","serial_number":"DXJG7890AB","asset_tag":"IT-003","category_id":1,"status":"active","is_up":true,"os":"Proxmox VE 8.1","has_rmm":true,"has_antivirus":false,"supplier_id":2,"installation_date":"2024-01-10","notes":"Virtualization host"}'
# Berpa HQ firewall
post /devices '{"site_id":1,"location_id":1,"model_id":8,"hostname":"FW-BERPA-01","dns_name":"fw-berpa-01.berpa.local","serial_number":"FG60F12345","asset_tag":"IT-010","category_id":6,"status":"active","is_up":true,"os":"FortiOS 7.4.2","has_rmm":false,"has_antivirus":false,"supplier_id":1,"installation_date":"2023-06-15","notes":"Edge firewall"}'
# Berpa HQ workstations
post /devices '{"site_id":1,"location_id":2,"model_id":3,"hostname":"PC-DIR-01","dns_name":"pc-dir-01.berpa.local","serial_number":"DELL90001","asset_tag":"IT-020","category_id":2,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-03-01","notes":"Director office"}'
post /devices '{"site_id":1,"location_id":3,"model_id":3,"hostname":"PC-OPEN-01","dns_name":"pc-open-01.berpa.local","serial_number":"DELL90002","asset_tag":"IT-021","category_id":2,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-03-01"}'
post /devices '{"site_id":1,"location_id":3,"model_id":3,"hostname":"PC-OPEN-02","dns_name":"pc-open-02.berpa.local","serial_number":"DELL90003","asset_tag":"IT-022","category_id":2,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-03-01"}'
post /devices '{"site_id":1,"location_id":3,"model_id":3,"hostname":"PC-OPEN-03","dns_name":"pc-open-03.berpa.local","serial_number":"DELL90004","asset_tag":"IT-023","category_id":2,"status":"active","is_up":false,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-03-01","notes":"Monitor issue, ticket #1234"}'
# Berpa HQ notebook
post /devices '{"site_id":1,"model_id":4,"hostname":"NB-ADMIN-01","dns_name":"nb-admin-01.berpa.local","serial_number":"LEN80001","asset_tag":"IT-030","category_id":3,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-06-01","notes":"IT admin laptop"}'
# Berpa HQ peripherals
post /devices '{"site_id":1,"location_id":3,"model_id":10,"hostname":"PRT-OPEN-01","serial_number":"HPP40001","asset_tag":"IT-040","category_id":8,"status":"active","is_up":true,"has_rmm":false,"has_antivirus":false,"supplier_id":1,"installation_date":"2024-03-01","notes":"Open space printer"}'
post /devices '{"site_id":1,"location_id":1,"model_id":9,"hostname":"UPS-RACK-01","serial_number":"APC50001","asset_tag":"IT-050","category_id":7,"status":"active","is_up":true,"has_rmm":false,"has_antivirus":false,"supplier_id":1,"installation_date":"2023-06-15","notes":"Server rack UPS"}'
# Berpa HQ AP
post /devices '{"site_id":1,"location_id":3,"model_id":7,"hostname":"AP-FLOOR1-01","serial_number":"UBQ60001","asset_tag":"IT-060","category_id":5,"status":"active","is_up":true,"has_rmm":false,"has_antivirus":false,"supplier_id":2,"installation_date":"2024-01-15"}'
# Berpa Cantiere
post /devices '{"site_id":2,"model_id":8,"hostname":"FW-CANT-01","serial_number":"FG60F22222","asset_tag":"IT-070","category_id":6,"status":"active","is_up":true,"os":"FortiOS 7.4.2","has_rmm":false,"has_antivirus":false,"supplier_id":1,"installation_date":"2024-06-01","notes":"Site-to-site VPN to HQ"}'
# OMP
post /devices '{"site_id":3,"location_id":4,"model_id":1,"hostname":"SRV-OMP-01","dns_name":"srv-omp-01.omp.local","serial_number":"CZJ55555AB","asset_tag":"OMP-001","category_id":1,"status":"active","is_up":true,"os":"Windows Server 2022","has_rmm":true,"has_antivirus":true,"supplier_id":1,"installation_date":"2023-09-01","notes":"File server + DC"}'
post /devices '{"site_id":3,"model_id":3,"hostname":"PC-OMP-01","dns_name":"pc-omp-01.omp.local","serial_number":"DELL70001","asset_tag":"OMP-010","category_id":2,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-02-01"}'
post /devices '{"site_id":3,"model_id":3,"hostname":"PC-OMP-02","dns_name":"pc-omp-02.omp.local","serial_number":"DELL70002","asset_tag":"OMP-011","category_id":2,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-02-01"}'
# Studio Rossi
post /devices '{"site_id":4,"location_id":5,"model_id":3,"hostname":"PC-ROSSI-01","serial_number":"DELL80001","asset_tag":"SLR-001","category_id":2,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-01-15","notes":"Avvocato 1"}'
post /devices '{"site_id":4,"location_id":5,"model_id":3,"hostname":"PC-ROSSI-02","serial_number":"DELL80002","asset_tag":"SLR-002","category_id":2,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-01-15","notes":"Avvocato 2"}'
post /devices '{"site_id":4,"location_id":5,"model_id":4,"hostname":"NB-ROSSI-01","serial_number":"LEN88001","asset_tag":"SLR-003","category_id":3,"status":"active","is_up":true,"os":"Windows 11 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2024-06-01","notes":"Mobile"}'
# Farmacia 1
post /devices '{"site_id":5,"location_id":6,"model_id":3,"hostname":"PC-FARM1-01","serial_number":"DELL99001","asset_tag":"FRC-001","category_id":2,"status":"active","is_up":true,"os":"Windows 10 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2022-04-01","notes":"POS/cash register"}'
post /devices '{"site_id":5,"location_id":6,"model_id":3,"hostname":"PC-FARM1-02","serial_number":"DELL99002","asset_tag":"FRC-002","category_id":2,"status":"active","is_up":true,"os":"Windows 10 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2022-04-01","notes":"Back office"}'
# Farmacia 2
post /devices '{"site_id":6,"model_id":3,"hostname":"PC-FARM2-01","serial_number":"DELL99003","asset_tag":"FRC-010","category_id":2,"status":"active","is_up":true,"os":"Windows 10 Pro","has_rmm":true,"has_antivirus":true,"supplier_id":2,"installation_date":"2022-04-01","notes":"POS/cash register"}'
# Decommissioned / storage
post /devices '{"site_id":1,"hostname":"SRV-OLD-01","serial_number":"OLD00001","asset_tag":"IT-900","category_id":1,"status":"decommissioned","is_up":false,"os":"Windows Server 2012 R2","has_rmm":false,"has_antivirus":false,"installation_date":"2018-01-01","notes":"Old DC, decommissioned 2023"}'
post /devices '{"site_id":1,"model_id":3,"hostname":"PC-SPARE-01","serial_number":"DELL00099","asset_tag":"IT-901","category_id":2,"status":"storage","is_up":false,"os":"Windows 11 Pro","has_rmm":false,"has_antivirus":false,"supplier_id":2,"installation_date":"2024-03-01","is_reserved":true,"notes":"Spare workstation"}'

# ── Device Interfaces ──
echo "==> Device Interfaces..."
# SRV-DC01 (device 1)
post /device-interfaces '{"device_id":1,"name":"eth0","mac_address":"00:11:22:33:44:01","notes":"Primary NIC"}'
post /device-interfaces '{"device_id":1,"name":"iDRAC","mac_address":"00:11:22:33:44:02","notes":"Management"}'
# SRV-DC02 (device 2)
post /device-interfaces '{"device_id":2,"name":"eth0","mac_address":"00:11:22:33:44:03"}'
post /device-interfaces '{"device_id":2,"name":"iDRAC","mac_address":"00:11:22:33:44:04"}'
# SRV-PROX01 (device 3)
post /device-interfaces '{"device_id":3,"name":"eno1","mac_address":"00:11:22:33:44:05","notes":"VM traffic"}'
post /device-interfaces '{"device_id":3,"name":"eno2","mac_address":"00:11:22:33:44:06","notes":"Storage network"}'
post /device-interfaces '{"device_id":3,"name":"iDRAC","mac_address":"00:11:22:33:44:07"}'
# FW-BERPA-01 (device 4)
post /device-interfaces '{"device_id":4,"name":"WAN","mac_address":"AA:BB:CC:DD:EE:01","notes":"ISP uplink"}'
post /device-interfaces '{"device_id":4,"name":"LAN","mac_address":"AA:BB:CC:DD:EE:02","notes":"Internal"}'
# PC-DIR-01 (device 5)
post /device-interfaces '{"device_id":5,"name":"eth0","mac_address":"00:22:33:44:55:01"}'
# PC-OPEN-01 (device 6)
post /device-interfaces '{"device_id":6,"name":"eth0","mac_address":"00:22:33:44:55:02"}'
# PC-OPEN-02 (device 7)
post /device-interfaces '{"device_id":7,"name":"eth0","mac_address":"00:22:33:44:55:03"}'
# PC-OPEN-03 (device 8)
post /device-interfaces '{"device_id":8,"name":"eth0","mac_address":"00:22:33:44:55:04"}'
# NB-ADMIN-01 (device 9)
post /device-interfaces '{"device_id":9,"name":"eth0","mac_address":"00:33:44:55:66:01","notes":"Docking station"}'
post /device-interfaces '{"device_id":9,"name":"wlan0","mac_address":"00:33:44:55:66:02","notes":"WiFi"}'
# AP-FLOOR1-01 (device 12)
post /device-interfaces '{"device_id":12,"name":"eth0","mac_address":"00:44:55:66:77:01","notes":"PoE"}'
# SRV-OMP-01 (device 14)
post /device-interfaces '{"device_id":14,"name":"eth0","mac_address":"00:55:66:77:88:01"}'
post /device-interfaces '{"device_id":14,"name":"iDRAC","mac_address":"00:55:66:77:88:02"}'
# PC-OMP-01 (device 15)
post /device-interfaces '{"device_id":15,"name":"eth0","mac_address":"00:55:66:77:88:03"}'
# PC-ROSSI-01 (device 17)
post /device-interfaces '{"device_id":17,"name":"eth0","mac_address":"00:66:77:88:99:01"}'
# PC-FARM1-01 (device 20)
post /device-interfaces '{"device_id":20,"name":"eth0","mac_address":"00:77:88:99:AA:01"}'

# ── Device IPs ──
echo "==> Device IPs..."
# SRV-DC01 eth0 (iface 1) → Servers VLAN (2)
post /device-ips '{"interface_id":1,"ip_address":"10.10.1.10","vlan_id":2,"is_primary":true,"notes":"Primary IP"}'
# SRV-DC01 iDRAC (iface 2) → Management VLAN (1)
post /device-ips '{"interface_id":2,"ip_address":"10.10.0.20","vlan_id":1,"is_primary":false,"notes":"iDRAC management"}'
# SRV-DC02 eth0 (iface 3)
post /device-ips '{"interface_id":3,"ip_address":"10.10.1.11","vlan_id":2,"is_primary":true}'
# SRV-DC02 iDRAC (iface 4)
post /device-ips '{"interface_id":4,"ip_address":"10.10.0.21","vlan_id":1,"is_primary":false}'
# SRV-PROX01 eno1 (iface 5)
post /device-ips '{"interface_id":5,"ip_address":"10.10.1.12","vlan_id":2,"is_primary":true}'
# SRV-PROX01 iDRAC (iface 7)
post /device-ips '{"interface_id":7,"ip_address":"10.10.0.22","vlan_id":1,"is_primary":false}'
# FW-BERPA-01 LAN (iface 9)
post /device-ips '{"interface_id":9,"ip_address":"10.10.0.1","vlan_id":1,"is_primary":true,"notes":"Default gateway"}'
# PC-DIR-01 (iface 10) → Users VLAN (3)
post /device-ips '{"interface_id":10,"ip_address":"10.10.2.10","vlan_id":3,"is_primary":true}'
# PC-OPEN-01 (iface 11)
post /device-ips '{"interface_id":11,"ip_address":"10.10.2.11","vlan_id":3,"is_primary":true}'
# PC-OPEN-02 (iface 12)
post /device-ips '{"interface_id":12,"ip_address":"10.10.2.12","vlan_id":3,"is_primary":true}'
# PC-OPEN-03 (iface 13)
post /device-ips '{"interface_id":13,"ip_address":"10.10.2.13","vlan_id":3,"is_primary":true}'
# NB-ADMIN-01 eth0 (iface 14)
post /device-ips '{"interface_id":14,"ip_address":"10.10.2.50","vlan_id":3,"is_primary":true}'
# AP-FLOOR1-01 (iface 16)
post /device-ips '{"interface_id":16,"ip_address":"10.10.0.30","vlan_id":1,"is_primary":true}'
# SRV-OMP-01 eth0 (iface 17) → OMP Servers VLAN (8)
post /device-ips '{"interface_id":17,"ip_address":"10.20.1.10","vlan_id":8,"is_primary":true}'
# SRV-OMP-01 iDRAC (iface 18) → OMP Management VLAN (7)
post /device-ips '{"interface_id":18,"ip_address":"10.20.0.20","vlan_id":7,"is_primary":false}'
# PC-OMP-01 (iface 19) → OMP Office VLAN (9)
post /device-ips '{"interface_id":19,"ip_address":"10.20.2.10","vlan_id":9,"is_primary":true}'
# PC-ROSSI-01 (iface 20) → Rossi LAN (11)
post /device-ips '{"interface_id":20,"ip_address":"192.168.1.10","vlan_id":11,"is_primary":true}'
# PC-FARM1-01 (iface 21) → Farmacia 1 LAN (12)
post /device-ips '{"interface_id":21,"ip_address":"192.168.10.10","vlan_id":12,"is_primary":true}'

# ── Device Connections ──
echo "==> Device Connections..."
# SRV-DC01 eth0 (iface 1) → SW-CORE-01 port 1 (sp 1) via PP-RACK-A-1 port 1 (pp 1)
post /device-connections '{"interface_id":1,"switch_port_id":1,"patch_panel_port_id":1,"connected_at":"2023-06-15"}'
# SRV-DC02 eth0 (iface 3) → SW-CORE-01 port 2 (sp 2) via PP-RACK-A-1 port 2 (pp 2)
post /device-connections '{"interface_id":3,"switch_port_id":2,"patch_panel_port_id":2,"connected_at":"2023-06-15"}'
# SRV-PROX01 eno1 (iface 5) → SW-CORE-01 port 3 (sp 3) via PP-RACK-A-1 port 3 (pp 3)
post /device-connections '{"interface_id":5,"switch_port_id":3,"patch_panel_port_id":3,"connected_at":"2024-01-10"}'
# PC-DIR-01 (iface 10) → SW-FLOOR1-01 port 1 (sp 49) via PP-RACK-A-2 port 1 (pp 25)
post /device-connections '{"interface_id":10,"switch_port_id":49,"patch_panel_port_id":25,"connected_at":"2024-03-01","notes":"Director office drop"}'
# PC-OPEN-01 (iface 11) → SW-FLOOR1-01 port 2 (sp 50) via PP-RACK-A-2 port 2 (pp 26)
post /device-connections '{"interface_id":11,"switch_port_id":50,"patch_panel_port_id":26,"connected_at":"2024-03-01"}'
# AP-FLOOR1-01 (iface 16) → SW-CORE-01 port 5 (sp 5) direct
post /device-connections '{"interface_id":16,"switch_port_id":5,"connected_at":"2024-01-15","notes":"PoE direct run"}'

echo ""
echo "==> Done! Database seeded successfully."
