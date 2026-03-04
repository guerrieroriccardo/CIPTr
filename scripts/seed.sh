#!/usr/bin/env bash
set -euo pipefail

DATABASE_URL="${DATABASE_URL:-postgres://ciptr:ciptr@172.17.4.23:5432/ciptr?sslmode=disable}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCHEMA="$SCRIPT_DIR/../backend/db/schema.sql"
SEED="$SCRIPT_DIR/../backend/db/seed.sql"

echo "==> Target: $DATABASE_URL"

# Ask for confirmation before wiping
read -rp "This will DROP all tables and re-seed. Continue? [y/N] " confirm
if [[ "$confirm" != [yY] ]]; then
    echo "Aborted."
    exit 0
fi

echo "==> Dropping all tables..."
psql "$DATABASE_URL" -q <<'SQL'
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
SQL

echo "==> Running schema.sql..."
psql "$DATABASE_URL" -q -f "$SCHEMA"

echo "==> Running seed.sql..."
psql "$DATABASE_URL" -q -f "$SEED"

echo "==> Done! Database seeded successfully."

# Quick summary
psql "$DATABASE_URL" -t -c "
SELECT 'clients: '     || count(*) FROM clients     UNION ALL
SELECT 'sites: '       || count(*) FROM sites       UNION ALL
SELECT 'devices: '     || count(*) FROM devices     UNION ALL
SELECT 'interfaces: '  || count(*) FROM device_interfaces UNION ALL
SELECT 'IPs: '         || count(*) FROM device_ips   UNION ALL
SELECT 'VLANs: '       || count(*) FROM vlans        UNION ALL
SELECT 'switches: '    || count(*) FROM switches     UNION ALL
SELECT 'connections: '  || count(*) FROM device_connections
ORDER BY 1;
"
