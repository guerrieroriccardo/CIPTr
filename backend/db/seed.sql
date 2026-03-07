-- ============================================================
-- SEED DATA — realistic Italian MSP scenario
-- Run after schema.sql to populate the database with sample data.
-- ============================================================

-- Clients
INSERT INTO clients (name, short_code, domain, notes) VALUES
  ('Berpa Costruzioni', 'BRP', 'berpa.local', 'Construction company'),
  ('Officine Meccaniche Pontina', 'OMP', 'omp.local', 'Mechanical workshop'),
  ('Studio Legale Rossi', 'SLR', NULL, 'Law firm'),
  ('Farmacia Centrale', 'FRC', NULL, 'Pharmacy chain');

-- Sites
INSERT INTO sites (client_id, name, address, notes) VALUES
  (1, 'Sede Principale', 'Via Roma 15, Sassuolo (MO)', 'HQ'),
  (1, 'Cantiere Nord', 'Via Emilia 200, Modena', 'Construction site office'),
  (2, 'Stabilimento', 'Via Pontina km 42, Latina', NULL),
  (3, 'Ufficio Centro', 'Corso Italia 8, Sassuolo (MO)', NULL),
  (4, 'Farmacia 1', 'Piazza Garibaldi 3, Sassuolo (MO)', NULL),
  (4, 'Farmacia 2', 'Via Radici 120, Sassuolo (MO)', NULL);

-- Locations
INSERT INTO locations (site_id, name, floor, notes) VALUES
  (1, 'Sala Server', '0', 'Basement server room'),
  (1, 'Ufficio Direzione', '1', NULL),
  (1, 'Open Space', '1', NULL),
  (3, 'Sala CED', '0', 'Small rack room'),
  (4, 'Ufficio', '0', NULL),
  (5, 'Retro Banco', '0', 'Behind the counter');

-- Manufacturers
INSERT INTO manufacturers (name) VALUES
  ('HP'),
  ('Cisco'),
  ('Dell'),
  ('Ubiquiti'),
  ('APC'),
  ('Lenovo'),
  ('Fortinet');

-- Categories
INSERT INTO categories (name, short_code) VALUES
  ('Server', 'SRV'),
  ('PC Desktop', 'PC'),
  ('Notebook', 'NB'),
  ('Switch', 'SW'),
  ('Access Point', 'AP'),
  ('Firewall', 'FW'),
  ('UPS', 'UPS'),
  ('Printer', 'PRT'),
  ('NAS', 'NAS'),
  ('Router', 'RTR');

-- Suppliers
INSERT INTO suppliers (name, address, phone, email) VALUES
  ('TechDistribution Srl', 'Via dell''Industria 5, Bologna', '+39 051 1234567', 'ordini@techdist.it'),
  ('InfoStore SpA', 'Via Giardini 300, Modena', '+39 059 9876543', 'vendite@infostore.it');

-- Device Models
INSERT INTO device_models (manufacturer_id, model_name, category_id, os_default, specs, notes) VALUES
  (1, 'ProLiant DL380 Gen10', 1, 'Windows Server 2022', '2x Xeon Gold, 64GB RAM, 4x 1.2TB SAS', NULL),
  (1, 'ProLiant DL360 Gen10', 1, 'Windows Server 2022', 'Xeon Silver, 32GB RAM, 2x 480GB SSD', NULL),
  (3, 'OptiPlex 7090', 2, 'Windows 11 Pro', 'i7-11700, 16GB RAM, 512GB NVMe', NULL),
  (6, 'ThinkPad T14s Gen3', 3, 'Windows 11 Pro', 'i7-1260P, 16GB RAM, 512GB NVMe', NULL),
  (2, 'Catalyst 2960-X 48', 4, NULL, '48x 1GbE, 4x SFP+', NULL),
  (2, 'Catalyst 2960-X 24', 4, NULL, '24x 1GbE, 4x SFP+', NULL),
  (4, 'UniFi U6 Pro', 5, NULL, 'Wi-Fi 6, PoE', NULL),
  (7, 'FortiGate 60F', 6, 'FortiOS 7.4', '10 GbE ports, 700 Mbps throughput', NULL),
  (5, 'Smart-UPS 1500', 7, NULL, '1500VA / 1000W, LCD, rack 2U', NULL),
  (1, 'LaserJet Pro M404dn', 8, NULL, 'B&W, duplex, network', NULL),
  (3, 'PowerEdge R640', 1, 'Proxmox VE 8', '2x Xeon Gold, 128GB RAM, 8x 960GB SSD', 'Virtualization host');

-- Address Blocks
INSERT INTO address_blocks (site_id, network, description, notes) VALUES
  (1, '10.10.0.0/20', 'Berpa HQ main block', NULL),
  (3, '10.20.0.0/22', 'OMP factory block', NULL),
  (4, '192.168.1.0/24', 'Studio Rossi single subnet', NULL),
  (5, '192.168.10.0/24', 'Farmacia 1', NULL),
  (6, '192.168.11.0/24', 'Farmacia 2', NULL);

-- VLANs
INSERT INTO vlans (site_id, address_block_id, vlan_id, name, subnet, description) VALUES
  -- Berpa HQ
  (1, 1, 1,   'Management',    '10.10.0.0/24',  'Network devices management'),
  (1, 1, 10,  'Servers',       '10.10.1.0/24',  'Server VLAN'),
  (1, 1, 20,  'Users',         '10.10.2.0/24',  'Workstations'),
  (1, 1, 30,  'VoIP',          '10.10.3.0/24',  'IP phones'),
  (1, 1, 40,  'Guest WiFi',    '10.10.4.0/24',  'Guest wireless network'),
  (1, 1, 99,  'Printers',      '10.10.9.0/24',  'Printers and MFPs'),
  -- OMP factory
  (3, 2, 1,   'Management',    '10.20.0.0/24',  NULL),
  (3, 2, 10,  'Servers',       '10.20.1.0/24',  NULL),
  (3, 2, 20,  'Office',        '10.20.2.0/24',  'Office workstations'),
  (3, 2, 30,  'Production',    '10.20.3.0/24',  'Factory floor devices'),
  -- Studio Rossi (flat)
  (4, 3, 1,   'LAN',           '192.168.1.0/24', 'Single flat network'),
  -- Farmacie
  (5, 4, 1,   'LAN',           '192.168.10.0/24', NULL),
  (6, 5, 1,   'LAN',           '192.168.11.0/24', NULL);

-- Switches
INSERT INTO switches (site_id, hostname, model_id, ip_address, location_id, total_ports, notes) VALUES
  (1, 'SW001', 5, '10.10.0.10', 1, 48, 'Core switch'),
  (1, 'SW002', 6, '10.10.0.11', 1, 24, 'Floor 1 access switch'),
  (3, 'SW001', 6, '10.20.0.10', 4, 24, NULL),
  (4, 'SW001', 6, '192.168.1.2', 5, 24, NULL);

-- Switch Ports (auto-created in real usage, but seed some for completeness)
-- SW-CORE-01 (48 ports)
INSERT INTO switch_ports (switch_id, port_number, port_label, speed, is_uplink) VALUES
  (1, 1, 'Gi1/0/1', '1G', false),
  (1, 2, 'Gi1/0/2', '1G', false),
  (1, 3, 'Gi1/0/3', '1G', false),
  (1, 4, 'Gi1/0/4', '1G', false),
  (1, 5, 'Gi1/0/5', '1G', false),
  (1, 6, 'Gi1/0/6', '1G', false),
  (1, 7, 'Gi1/0/7', '1G', false),
  (1, 8, 'Gi1/0/8', '1G', false),
  (1, 47, 'Gi1/0/47', '10G', true),
  (1, 48, 'Gi1/0/48', '10G', true);
-- SW-FLOOR1-01 (24 ports)
INSERT INTO switch_ports (switch_id, port_number, port_label, speed, is_uplink) VALUES
  (2, 1, 'Gi0/1', '1G', false),
  (2, 2, 'Gi0/2', '1G', false),
  (2, 3, 'Gi0/3', '1G', false),
  (2, 4, 'Gi0/4', '1G', false),
  (2, 23, 'Gi0/23', '1G', true),
  (2, 24, 'Gi0/24', '1G', true);

-- Patch Panels
INSERT INTO patch_panels (site_id, name, total_ports, location, notes) VALUES
  (1, 'PP-RACK-A-1', 24, 'Sala Server - Rack A - 1U', 'Top of rack'),
  (1, 'PP-RACK-A-2', 24, 'Sala Server - Rack A - 2U', NULL);

-- Patch Panel Ports (sample)
INSERT INTO patch_panel_ports (patch_panel_id, port_number, port_label) VALUES
  (1, 1, 'A1-01'), (1, 2, 'A1-02'), (1, 3, 'A1-03'), (1, 4, 'A1-04'),
  (1, 5, 'A1-05'), (1, 6, 'A1-06'), (1, 7, 'A1-07'), (1, 8, 'A1-08'),
  (2, 1, 'A2-01'), (2, 2, 'A2-02'), (2, 3, 'A2-03'), (2, 4, 'A2-04');

-- Devices
INSERT INTO devices (site_id, location_id, model_id, hostname, dns_name, serial_number, asset_tag, category_id, status, is_up, os, has_rmm, has_antivirus, supplier_id, installation_date, is_reserved, notes) VALUES
  -- Berpa HQ servers
  (1, 1, 1, 'SRV-DC01',    'srv-dc01.berpa.local',    'CZJ12345AB', 'IT-001', 1, 'active', true,  'Windows Server 2022', true, true, 1, '2023-06-15', false, 'Primary domain controller'),
  (1, 1, 2, 'SRV-DC02',    'srv-dc02.berpa.local',    'CZJ12345AC', 'IT-002', 1, 'active', true,  'Windows Server 2022', true, true, 1, '2023-06-15', false, 'Secondary domain controller'),
  (1, 1, 11, 'SRV-PROX01', 'srv-prox01.berpa.local',  'DXJG7890AB', 'IT-003', 1, 'active', true,  'Proxmox VE 8.1',      true, false, 2, '2024-01-10', false, 'Virtualization host'),
  -- Berpa HQ firewall
  (1, 1, 8, 'FW-BERPA-01', 'fw-berpa-01.berpa.local', 'FG60F12345', 'IT-010', 6, 'active', true,  'FortiOS 7.4.2',       false, false, 1, '2023-06-15', false, 'Edge firewall'),
  -- Berpa HQ workstations
  (1, 2, 3, 'PC-DIR-01',   'pc-dir-01.berpa.local',   'DELL90001', 'IT-020', 2, 'active', true,  'Windows 11 Pro',      true, true, 2, '2024-03-01', false, 'Director office'),
  (1, 3, 3, 'PC-OPEN-01',  'pc-open-01.berpa.local',  'DELL90002', 'IT-021', 2, 'active', true,  'Windows 11 Pro',      true, true, 2, '2024-03-01', false, NULL),
  (1, 3, 3, 'PC-OPEN-02',  'pc-open-02.berpa.local',  'DELL90003', 'IT-022', 2, 'active', true,  'Windows 11 Pro',      true, true, 2, '2024-03-01', false, NULL),
  (1, 3, 3, 'PC-OPEN-03',  'pc-open-03.berpa.local',  'DELL90004', 'IT-023', 2, 'active', false, 'Windows 11 Pro',      true, true, 2, '2024-03-01', false, 'Monitor issue, ticket #1234'),
  -- Berpa HQ notebooks
  (1, NULL, 4, 'NB-ADMIN-01', 'nb-admin-01.berpa.local', 'LEN80001', 'IT-030', 3, 'active', true, 'Windows 11 Pro', true, true, 2, '2024-06-01', false, 'IT admin laptop'),
  -- Berpa HQ peripherals
  (1, 3, 10, 'PRT-OPEN-01', NULL, 'HPP40001', 'IT-040', 8, 'active', true, NULL, false, false, 1, '2024-03-01', false, 'Open space printer'),
  (1, 1, 9,  'UPS-RACK-01', NULL, 'APC50001', 'IT-050', 7, 'active', true, NULL, false, false, 1, '2023-06-15', false, 'Server rack UPS'),
  -- Berpa HQ AP
  (1, 3, 7, 'AP-FLOOR1-01', NULL, 'UBQ60001', 'IT-060', 5, 'active', true, NULL, false, false, 2, '2024-01-15', false, NULL),
  -- Berpa Cantiere
  (2, NULL, 8, 'FW-CANT-01', NULL, 'FG60F22222', 'IT-070', 6, 'active', true, 'FortiOS 7.4.2', false, false, 1, '2024-06-01', false, 'Site-to-site VPN to HQ'),
  -- OMP
  (3, 4, 1, 'SRV-OMP-01',  'srv-omp-01.omp.local',   'CZJ55555AB', 'OMP-001', 1, 'active', true,  'Windows Server 2022', true, true, 1, '2023-09-01', false, 'File server + DC'),
  (3, NULL, 3, 'PC-OMP-01', 'pc-omp-01.omp.local',    'DELL70001',  'OMP-010', 2, 'active', true,  'Windows 11 Pro',      true, true, 2, '2024-02-01', false, NULL),
  (3, NULL, 3, 'PC-OMP-02', 'pc-omp-02.omp.local',    'DELL70002',  'OMP-011', 2, 'active', true,  'Windows 11 Pro',      true, true, 2, '2024-02-01', false, NULL),
  -- Studio Rossi
  (4, 5, 3, 'PC-ROSSI-01', NULL, 'DELL80001', 'SLR-001', 2, 'active', true, 'Windows 11 Pro', true, true, 2, '2024-01-15', false, 'Avvocato 1'),
  (4, 5, 3, 'PC-ROSSI-02', NULL, 'DELL80002', 'SLR-002', 2, 'active', true, 'Windows 11 Pro', true, true, 2, '2024-01-15', false, 'Avvocato 2'),
  (4, 5, 4, 'NB-ROSSI-01', NULL, 'LEN88001', 'SLR-003', 3, 'active', true, 'Windows 11 Pro', true, true, 2, '2024-06-01', false, 'Mobile'),
  -- Farmacia 1
  (5, 6, 3, 'PC-FARM1-01', NULL, 'DELL99001', 'FRC-001', 2, 'active', true, 'Windows 10 Pro', true, true, 2, '2022-04-01', false, 'POS/cash register'),
  (5, 6, 3, 'PC-FARM1-02', NULL, 'DELL99002', 'FRC-002', 2, 'active', true, 'Windows 10 Pro', true, true, 2, '2022-04-01', false, 'Back office'),
  -- Farmacia 2
  (6, NULL, 3, 'PC-FARM2-01', NULL, 'DELL99003', 'FRC-010', 2, 'active', true, 'Windows 10 Pro', true, true, 2, '2022-04-01', false, 'POS/cash register'),
  -- Decommissioned / storage
  (1, NULL, NULL, 'SRV-OLD-01', NULL, 'OLD00001', 'IT-900', 1, 'decommissioned', false, 'Windows Server 2012 R2', false, false, NULL, '2018-01-01', false, 'Old DC, decommissioned 2023'),
  (1, NULL, 3, 'PC-SPARE-01', NULL, 'DELL00099', 'IT-901', 2, 'storage', false, 'Windows 11 Pro', false, false, 2, '2024-03-01', true, 'Spare workstation');

-- Device Interfaces
INSERT INTO device_interfaces (device_id, name, mac_address, notes) VALUES
  -- SRV-DC01
  (1, 'eth0', '00:11:22:33:44:01', 'Primary NIC'),
  (1, 'iDRAC', '00:11:22:33:44:02', 'Management'),
  -- SRV-DC02
  (2, 'eth0', '00:11:22:33:44:03', NULL),
  (2, 'iDRAC', '00:11:22:33:44:04', NULL),
  -- SRV-PROX01
  (3, 'eno1', '00:11:22:33:44:05', 'VM traffic'),
  (3, 'eno2', '00:11:22:33:44:06', 'Storage network'),
  (3, 'iDRAC', '00:11:22:33:44:07', NULL),
  -- FW-BERPA-01
  (4, 'WAN', 'AA:BB:CC:DD:EE:01', 'ISP uplink'),
  (4, 'LAN', 'AA:BB:CC:DD:EE:02', 'Internal'),
  -- PC-DIR-01
  (5, 'eth0', '00:22:33:44:55:01', NULL),
  -- PC-OPEN-01
  (6, 'eth0', '00:22:33:44:55:02', NULL),
  -- PC-OPEN-02
  (7, 'eth0', '00:22:33:44:55:03', NULL),
  -- PC-OPEN-03
  (8, 'eth0', '00:22:33:44:55:04', NULL),
  -- NB-ADMIN-01
  (9, 'eth0', '00:33:44:55:66:01', 'Docking station'),
  (9, 'wlan0', '00:33:44:55:66:02', 'WiFi'),
  -- AP-FLOOR1-01
  (12, 'eth0', '00:44:55:66:77:01', 'PoE'),
  -- SRV-OMP-01
  (14, 'eth0', '00:55:66:77:88:01', NULL),
  (14, 'iDRAC', '00:55:66:77:88:02', NULL),
  -- PC-OMP-01
  (15, 'eth0', '00:55:66:77:88:03', NULL),
  -- PC-ROSSI-01
  (17, 'eth0', '00:66:77:88:99:01', NULL),
  -- PC-FARM1-01
  (20, 'eth0', '00:77:88:99:AA:01', NULL);

-- Device IPs
INSERT INTO device_ips (interface_id, ip_address, vlan_id, is_primary, notes) VALUES
  -- SRV-DC01 eth0
  (1, '10.10.1.10', 2, true, 'Primary IP'),
  -- SRV-DC01 iDRAC
  (2, '10.10.0.20', 1, false, 'iDRAC management'),
  -- SRV-DC02 eth0
  (3, '10.10.1.11', 2, true, NULL),
  -- SRV-DC02 iDRAC
  (4, '10.10.0.21', 1, false, NULL),
  -- SRV-PROX01 eno1
  (5, '10.10.1.12', 2, true, NULL),
  -- SRV-PROX01 iDRAC
  (7, '10.10.0.22', 1, false, NULL),
  -- FW-BERPA-01 LAN
  (9, '10.10.0.1', 1, true, 'Default gateway'),
  -- PC-DIR-01
  (10, '10.10.2.10', 3, true, NULL),
  -- PC-OPEN-01
  (11, '10.10.2.11', 3, true, NULL),
  -- PC-OPEN-02
  (12, '10.10.2.12', 3, true, NULL),
  -- PC-OPEN-03
  (13, '10.10.2.13', 3, true, NULL),
  -- NB-ADMIN-01 eth0
  (14, '10.10.2.50', 3, true, NULL),
  -- AP-FLOOR1-01
  (16, '10.10.0.30', 1, true, NULL),
  -- SRV-OMP-01
  (17, '10.20.1.10', 8, true, NULL),
  -- SRV-OMP-01 iDRAC
  (18, '10.20.0.20', 7, false, NULL),
  -- PC-OMP-01
  (19, '10.20.2.10', 9, true, NULL),
  -- PC-ROSSI-01
  (20, '192.168.1.10', 11, true, NULL),
  -- PC-FARM1-01
  (21, '192.168.10.10', 12, true, NULL);

-- Device Connections (some sample cabling)
INSERT INTO device_connections (interface_id, switch_port_id, patch_panel_port_id, connected_at, notes) VALUES
  -- SRV-DC01 eth0 → SW-CORE-01 port 1 via PP-RACK-A-1 port 1
  (1, 1, 1, '2023-06-15', NULL),
  -- SRV-DC02 eth0 → SW-CORE-01 port 2 via PP-RACK-A-1 port 2
  (3, 2, 2, '2023-06-15', NULL),
  -- SRV-PROX01 eno1 → SW-CORE-01 port 3 via PP-RACK-A-1 port 3
  (5, 3, 3, '2024-01-10', NULL),
  -- PC-DIR-01 → SW-FLOOR1-01 port 1 via PP-RACK-A-2 port 1
  (10, 11, 9, '2024-03-01', 'Director office drop'),
  -- PC-OPEN-01 → SW-FLOOR1-01 port 2 via PP-RACK-A-2 port 2
  (11, 12, 10, '2024-03-01', NULL),
  -- AP-FLOOR1-01 → SW-CORE-01 port 5 (direct, no patch panel)
  (16, 5, NULL, '2024-01-15', 'PoE direct run');
