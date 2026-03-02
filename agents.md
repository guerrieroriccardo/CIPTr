# CIPTr (Client IP Tracker) — Specifiche per Agenti AI

Questo documento descrive la webapp **CIPTr** per la gestione degli asset di rete di un MSP.
Serve da guida completa per agenti AI (o sviluppatori) che devono costruire l'applicazione passo per passo.

---

## 1. Panoramica del Progetto

### Obiettivo
Sostituire un foglio Excel con una webapp che permette a un team MSP di:
- Tenere traccia di tutti i dispositivi nelle reti dei clienti
- Gestire IP, connessioni switch/patch-panel, VLAN
- Avere un inventario dei modelli hardware
- Vedere lo storico delle installazioni e i ticket correlati

### Stack Tecnico
| Layer | Tecnologia |
|-------|-----------|
| Backend | Go (Gin framework) |
| Database | PostgreSQL 18 (driver: `pgx/v5`) |
| Query | raw SQL con `database/sql` |
| CLI | Go (bubbletea TUI) |
| Frontend | React + TypeScript (da fare) |
| UI Components | shadcn/ui (basato su Tailwind CSS + Radix UI) |
| HTTP Client | TanStack Query (react-query) per fetching e caching |
| Routing | React Router v6 |
| Build Tool | Vite |
| Containerizzazione | Docker + Docker Compose |
| TLS/SSL | Gestito da reverse proxy esterno (es. Nginx, Traefik) |

### Repository e Moduli Go
- GitHub: `github.com/guerrieroriccardo/CIPTr`
- Go workspace (`go.work`) alla root collega `backend/` e `cli/`
- Modulo backend: `github.com/guerrieroriccardo/CIPTr/backend`
- Modulo CLI: `github.com/guerrieroriccardo/CIPTr/cli`
- La CLI importa `backend/models` via workspace (nessuna duplicazione struct)

### Note Docker
- Il backend gira in HTTP puro (porta 8080) — TLS è responsabilità del reverse proxy
- PostgreSQL persistito tramite Docker volume
- Connessione DB configurabile via env `DATABASE_URL`
- Build multi-stage per immagine backend minimale (golang:alpine → alpine)
- Il frontend viene servito da nginx:alpine con proxy `/api` → backend

### Vincoli
- Nessun sistema di autenticazione/permessi
- Semplicità prima di tutto
- Facilmente mantenibile da chi non è esperto di programmazione

---

## 2. Struttura del Progetto

```
CIPTr/
├── go.work                     # Go workspace (collega backend/ e cli/)
├── backend/
│   ├── main.go
│   ├── go.mod                  # module github.com/guerrieroriccardo/CIPTr/backend
│   ├── go.sum
│   ├── router.go               # Gin engine + routes
│   ├── db/
│   │   ├── schema.sql          # Schema PostgreSQL
│   │   └── database.go         # Inizializzazione DB
│   ├── handlers/               # Handler HTTP per ogni risorsa
│   │   ├── health.go
│   │   ├── response.go         # ok() e fail() helpers
│   │   ├── clients.go
│   │   ├── devices.go
│   │   ├── switches.go
│   │   └── ...
│   ├── models/                 # Struct Go che rispecchiano le tabelle (condivise con CLI)
│   └── Dockerfile              # Multi-stage build
├── cli/
│   ├── main.go                 # Entry point bubbletea
│   ├── go.mod                  # module github.com/guerrieroriccardo/CIPTr/cli
│   └── internal/
│       ├── apiclient/          # Client HTTP per la REST API
│       │   ├── client.go       # HTTP helpers + envelope parsing
│       │   └── clients.go      # Metodi per ogni risorsa (uno per file)
│       └── tui/                # Componenti bubbletea
│           ├── app.go          # Root model + dispatching
│           ├── nav.go          # Stack di navigazione + breadcrumb
│           ├── styles.go       # Stili lipgloss
│           ├── menu.go         # Menu principale
│           ├── table.go        # Tabella generica per ogni risorsa
│           ├── form.go         # Form generico create/edit
│           ├── confirm.go      # Dialog conferma eliminazione
│           └── resource/       # Definizioni per risorsa (colonne, campi form)
│               ├── registry.go # Tipo Def + registro
│               └── clients.go  # ClientDef, SiteDef, ecc.
├── frontend/
│   ├── src/
│   │   ├── api/                # Funzioni fetch verso il backend
│   │   ├── components/         # Componenti shadcn/ui personalizzati
│   │   ├── pages/              # Pagine principali
│   │   └── App.tsx
│   ├── nginx.conf              # Config nginx per servire il frontend
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile              # Build + nginx:alpine
├── docker-compose.yml          # Orchestrazione backend + frontend
├── .gitignore
└── agents.md                   # Questo file
```

---

## 3. Schema del Database

### Diagramma Logico

```
clients
  └── sites (physical locations of the client)
        └── address_blocks  (/20 block assigned to the site)
              └── vlans (subnets carved from the block, e.g. /24 per VLAN)
        └── locations (rooms, floors, closets within the site)
        └── switches (physical switches at the site)
              └── switch_ports (ports on each switch)
        └── patch_panels (patch panels at the site)
              └── patch_panel_ports (ports on each patch panel)

device_models (hardware catalog: HP ProLiant, Cisco 2960, ...)
  └── devices (deployed devices)
        └── device_interfaces (NICs: eth0, iDRAC, WAN, LAN1, ...)
              └── device_ips (IP per NIC, linked to a VLAN)
              └── device_connections (physical link: NIC → switch_port ↔ patch_panel_port)
```

### SQL — schema.sql

```sql
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

-- ============================================================
-- CLIENTS AND SITES
-- ============================================================

CREATE TABLE IF NOT EXISTS clients (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    short_code  TEXT NOT NULL UNIQUE,   -- e.g. "ADP", "XYZ"
    notes       TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sites (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id   INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,          -- e.g. "HQ", "Rome Branch"
    address     TEXT,
    notes       TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(client_id, name)
);

CREATE TABLE IF NOT EXISTS locations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id     INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,          -- e.g. "IT Dept", "Reception"
    floor       TEXT,
    notes       TEXT,
    UNIQUE(site_id, name)
);

-- ============================================================
-- IP ADDRESS SPACE
-- ============================================================

-- One /20 block (or any prefix) is assigned to each site.
-- Multiple blocks per site are allowed for flexibility.
CREATE TABLE IF NOT EXISTS address_blocks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id     INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    network     TEXT NOT NULL,          -- e.g. "10.10.0.0/20"
    description TEXT,
    notes       TEXT,
    UNIQUE(site_id, network)
);

-- VLANs are subnets carved from an address_block (e.g. /24 per VLAN)
CREATE TABLE IF NOT EXISTS vlans (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id          INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    address_block_id INTEGER REFERENCES address_blocks(id),
    vlan_id          INTEGER NOT NULL,  -- VLAN number, e.g. 10, 20, 100
    name             TEXT NOT NULL,     -- e.g. "Users LAN", "VOIP"
    subnet           TEXT,              -- e.g. "10.10.0.0/24"
    gateway          TEXT,
    description      TEXT,
    UNIQUE(site_id, vlan_id)
);

-- ============================================================
-- NETWORK INFRASTRUCTURE
-- ============================================================

CREATE TABLE IF NOT EXISTS switches (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id         INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,      -- e.g. "SW006", "CORE-SW01"
    model_id        INTEGER REFERENCES device_models(id),
    ip_address      TEXT,
    location        TEXT,               -- e.g. "Rack A, Cabinet 3"
    total_ports     INTEGER NOT NULL DEFAULT 24,
    notes           TEXT,
    UNIQUE(site_id, name)
);

CREATE TABLE IF NOT EXISTS switch_ports (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    switch_id       INTEGER NOT NULL REFERENCES switches(id) ON DELETE CASCADE,
    port_number     INTEGER NOT NULL,
    port_label      TEXT,               -- optional label
    speed           TEXT,               -- e.g. "1G", "10G"
    is_uplink       BOOLEAN DEFAULT 0,
    notes           TEXT,
    UNIQUE(switch_id, port_number)
);

CREATE TABLE IF NOT EXISTS patch_panels (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id         INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,      -- e.g. "PP-RACK1-A"
    total_ports     INTEGER NOT NULL DEFAULT 24,
    location        TEXT,
    notes           TEXT,
    UNIQUE(site_id, name)
);

CREATE TABLE IF NOT EXISTS patch_panel_ports (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    patch_panel_id      INTEGER NOT NULL REFERENCES patch_panels(id) ON DELETE CASCADE,
    port_number         INTEGER NOT NULL,
    port_label          TEXT,
    notes               TEXT,
    UNIQUE(patch_panel_id, port_number)
);

-- ============================================================
-- DEVICE CATALOG (INVENTORY)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_models (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    manufacturer    TEXT NOT NULL,      -- e.g. "HP", "Cisco", "Dell"
    model_name      TEXT NOT NULL,      -- e.g. "ProLiant DL360 Gen10"
    category        TEXT NOT NULL,      -- Server, PC, Laptop, Printer, Switch, Router, AP, NAS, Camera, Phone, UPS, Other
    os_default      TEXT,               -- typical OS for this model
    specs           TEXT,               -- free text: CPU, RAM, etc.
    notes           TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(manufacturer, model_name)
);

-- ============================================================
-- DEPLOYED DEVICES
-- ============================================================

CREATE TABLE IF NOT EXISTS devices (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id             INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    location_id           INTEGER REFERENCES locations(id),
    model_id            INTEGER REFERENCES device_models(id),

    -- Identification
    hostname            TEXT NOT NULL,
    dns_name            TEXT,
    serial_number       TEXT,
    asset_tag           TEXT,

    -- Type and status
    device_type         TEXT NOT NULL,  -- PC, Server, Printer, Switch, AP, Camera, Phone, NAS, UPS, Other
    status              TEXT NOT NULL DEFAULT 'active',  -- active, inactive, reserved, decommissioned
    is_up               BOOLEAN DEFAULT 1,

    -- Software / management
    os                  TEXT,
    has_rmm             BOOLEAN DEFAULT 0,  -- RMM agent installed
    has_antivirus       BOOLEAN DEFAULT 0,  -- antivirus installed
    supplier            TEXT,

    -- Logistics
    installation_date   DATE,
    is_reserved         BOOLEAN DEFAULT 0,

    -- Ticket / reason
    ticket_ref          TEXT,
    reason              TEXT,

    notes               TEXT,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Trigger to auto-update updated_at
CREATE TRIGGER IF NOT EXISTS devices_updated_at
AFTER UPDATE ON devices
BEGIN
    UPDATE devices SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- ============================================================
-- IP ADDRESSES (multiple IPs per device)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_ips (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id   INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    ip_address  TEXT NOT NULL,
    mac_address TEXT,
    vlan_id     INTEGER REFERENCES vlans(id),
    is_primary  BOOLEAN DEFAULT 0,
    interface   TEXT,                   -- e.g. "eth0", "Wi-Fi"
    notes       TEXT
);

-- ============================================================
-- PHYSICAL CONNECTIONS (switch port → device)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_connections (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id           INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    switch_port_id      INTEGER REFERENCES switch_ports(id),
    patch_panel_port_id INTEGER REFERENCES patch_panel_ports(id),
    connected_at        DATE,
    notes               TEXT
);

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_devices_site         ON devices(site_id);
CREATE INDEX IF NOT EXISTS idx_devices_hostname     ON devices(hostname);
CREATE INDEX IF NOT EXISTS idx_device_ips_address   ON device_ips(ip_address);
CREATE INDEX IF NOT EXISTS idx_switch_ports_sw      ON switch_ports(switch_id);
CREATE INDEX IF NOT EXISTS idx_address_blocks_site  ON address_blocks(site_id);
CREATE INDEX IF NOT EXISTS idx_vlans_block          ON vlans(address_block_id);
```

---

## 4. API REST del Backend (Go)

Base URL: `http://localhost:8080/api/v1`

### Clients
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/clients` | Lista tutti i clienti |
| POST | `/clients` | Crea cliente |
| GET | `/clients/:id` | Dettaglio cliente |
| PUT | `/clients/:id` | Aggiorna cliente |
| DELETE | `/clients/:id` | Elimina cliente |
| GET | `/clients/:id/sites` | Sedi del cliente |

### Sites
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/sites` | Lista sedi (con filtro ?client_id=) |
| POST | `/sites` | Crea sede |
| GET | `/sites/:id` | Dettaglio sede |
| PUT | `/sites/:id` | Aggiorna sede |
| DELETE | `/sites/:id` | Elimina sede |

### Devices
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/devices` | Lista dispositivi (filtri: site_id, status, device_type, search) |
| POST | `/devices` | Crea dispositivo |
| GET | `/devices/:id` | Dettaglio dispositivo (con IP e connessioni) |
| PUT | `/devices/:id` | Aggiorna dispositivo |
| DELETE | `/devices/:id` | Elimina dispositivo |
| GET | `/devices/:id/ips` | IP del dispositivo |
| POST | `/devices/:id/ips` | Aggiungi IP |
| DELETE | `/device-ips/:id` | Rimuovi IP |

### Switches
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/switches` | Lista switch (filtro ?site_id=) |
| POST | `/switches` | Crea switch |
| GET | `/switches/:id` | Dettaglio switch |
| PUT | `/switches/:id` | Aggiorna switch |
| DELETE | `/switches/:id` | Elimina switch |
| GET | `/switches/:id/ports` | Porte dello switch con dispositivo collegato |
| PUT | `/switch-ports/:id` | Aggiorna porta (collega/scollega dispositivo) |

### Device Models (Catalogo)
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/device-models` | Lista modelli |
| POST | `/device-models` | Crea modello |
| GET | `/device-models/:id` | Dettaglio modello |
| PUT | `/device-models/:id` | Aggiorna modello |
| DELETE | `/device-models/:id` | Elimina modello |

### VLANs
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/vlans` | Lista VLAN (filtro ?site_id=) |
| POST | `/vlans` | Crea VLAN |
| PUT | `/vlans/:id` | Aggiorna VLAN |
| DELETE | `/vlans/:id` | Elimina VLAN |

---

## 5. Frontend — Pagine e Navigazione

```
/                       → Dashboard (statistiche generali)
/clients                → Lista clienti
/clients/:id            → Dettaglio cliente (con lista sedi)
/clients/:id/sites/:sid → Dettaglio sede (reti, switch, dispositivi)
/devices                → Lista globale dispositivi (con filtri)
/devices/:id            → Dettaglio dispositivo
/devices/new            → Aggiungi dispositivo
/switches               → Lista switch
/switches/:id           → Dettaglio switch (visualizzazione porte)
/inventory              → Catalogo modelli hardware
/vlans                  → Gestione VLAN
```

### Componenti UI Principali

- **DataTable** — tabella con ordinamento, filtri, paginazione (shadcn/ui DataTable)
- **DeviceForm** — form per creare/modificare dispositivo
- **SwitchPortMap** — visualizzazione grafica delle porte di uno switch
- **IPManager** — gestione multipla degli IP di un dispositivo
- **StatusBadge** — badge colorato per stato dispositivo

---

## 6. Fasi di Sviluppo (per agente AI)

### Fase 1 — Setup Progetto ✅
1. Inizializzare modulo Go in `backend/`
2. Installare dipendenze Go: `gin`, `pgx/v5`
3. Creare `backend/db/schema.sql` e funzione `db.Open()`
4. Docker Compose con PostgreSQL

### Fase 2 — Backend Core ✅
Tutte le 14 risorse CRUD implementate e testate:
clients, sites, locations, address_blocks, vlans, device_models, devices,
device_interfaces, device_ips, device_connections, switches, switch_ports,
patch_panels, patch_panel_ports

### Fase 3 — CLI (bubbletea TUI) 🔄
TUI interattiva per gestire i dati via REST API (senza frontend web).
- Stack: bubbletea + bubbles + lipgloss
- Navigazione: gerarchica (Client → Site → risorse) + accesso flat a tutte le 14 risorse
- Componenti generici (table, form, confirm) guidati da un registro di definizioni per risorsa
- Passi:
  1. ✅ Scaffold CLI + Go workspace
  2. ✅ API client con envelope parsing
  3. ✅ Stack di navigazione + App root
  4. ✅ Menu principale
  5. 🔄 Registro risorse + definizione Clients
  6. API client per Clients
  7. Tabella generica
  8. Form generico + conferma eliminazione
  9. Sites + navigazione gerarchica
  10. Risorse rimanenti (una alla volta)
  11. Polish (statusbar, spinner, filtri)

### Fase 4 — Frontend Core
1. Inizializzare progetto React con Vite + TypeScript in `frontend/`
2. Installare e configurare shadcn/ui, TanStack Query, React Router
3. Creare funzioni API client in `src/api/` (una per risorsa)
4. Creare tipi TypeScript in `src/types/` speculari ai modelli Go
5. Implementare pagina Lista Clienti con DataTable
6. Implementare pagina Dettaglio Sede (tab: Dispositivi, Switch, VLAN)
7. Implementare pagina Lista Dispositivi con filtri
8. Implementare form creazione/modifica dispositivo

### Fase 5 — Features Avanzate
1. Implementare SwitchPortMap (griglia visiva delle porte)
2. Implementare catalogo modelli (Inventario) con CRUD
3. Implementare gestione VLAN
4. Aggiungere ricerca globale per hostname/IP
5. Dashboard con statistiche (contatori per cliente/stato)

### Fase 6 — Rifinitura
1. Toast per conferma operazioni (shadcn/ui Toaster)
2. Dialog di conferma per eliminazioni
3. Esportazione CSV della lista dispositivi
4. File `README.md` con istruzioni di avvio

---

## 7. Regole e Convenzioni

### Lingua
- **Tutto il codice sorgente è in inglese**: variabili, funzioni, commenti, messaggi di errore, nomi di file, commit message.
- Questo documento (`agents.md`) rimane in italiano perché è una specifica per il team.
- In futuro si potrà aggiungere l'i18n per l'interfaccia utente (testi delle pagine), ma non ora.

### Backend (Go)
- Handler restituiscono sempre JSON: `{"data": ..., "error": null}` oppure `{"data": null, "error": "messaggio"}`
- Codici HTTP standard: 200, 201, 400, 404, 500
- Tutti gli ID nelle URL sono interi
- Query parametri per filtri: `?site_id=1&status=active&search=testo`
- PostgreSQL con parametri `$1, $2, ...` e `RETURNING` per INSERT/UPDATE
- Tipi PostgreSQL: `MACADDR` per MAC address, `INET` per IP, `CIDR` per blocchi rete
- Connessione DB via env `DATABASE_URL` (default `postgres://ciptr:ciptr@localhost:5432/ciptr`)
- Server di default porta `8080` (configurabile via env `PORT`)
- Nessun framework ORM: SQL diretto con `database/sql`

### CLI (bubbletea)
- Binario separato in `cli/`, chiama la REST API (non accede al DB direttamente)
- URL API configurabile via env `CIPTR_API_URL` (default `http://localhost:8080/api/v1`)
- Componenti generici (table, form) guidati da un registro di definizioni (`resource/registry.go`)
- Navigazione stack-based: `PushScreenMsg` / `PopScreenMsg` / Esc per tornare indietro

### Frontend (React)
- Un file per pagina in `src/pages/`
- Un file per risorsa API in `src/api/` (es. `src/api/devices.ts`)
- Usare sempre TanStack Query per fetch (no fetch diretto nei componenti)
- Nessun global state manager
- Tutti i tipi TypeScript in `src/types/index.ts`

### Ritmo di lavoro
- **Una risorsa alla volta**: implementare model, handler e routes per UNA sola risorsa, poi committare e attendere review prima di procedere alla successiva
- Non accorpare più risorse in un unico blocco di lavoro
- L'utente vuole poter leggere e capire ogni cambiamento singolarmente
- **Aggiornare sempre `agents.md`** quando si pianificano cambiamenti architetturali, nuove feature, nuovi componenti o nuove fasi — questo file è la fonte di verità del progetto

### Git
- Committare spesso: dopo ogni risorsa CRUD completata, ogni refactor, ogni modifica significativa
- Non aggiungere `Co-Authored-By` nei commit message
- Commit message in inglese, stile conventional commits (`feat`, `fix`, `refactor`, ecc.)

### Database
- PostgreSQL 18 — FK enforcement è attivo di default
- Nessun soft delete: eliminazione reale con conferma nel frontend/CLI
- Schema in `backend/db/schema.sql`

---

## 8. Valori Enumerati

### devices.device_type
`PC`, `Laptop`, `Server`, `Printer`, `Switch`, `Router`, `AP`, `NAS`, `Camera`, `Phone`, `UPS`, `Other`

### devices.status
`active`, `inactive`, `reserved`, `decommissioned`

### device_models.category
`PC`, `Laptop`, `Server`, `Printer`, `Switch`, `Router`, `AP`, `NAS`, `Camera`, `Phone`, `UPS`, `Other`

---

## 9. Note Importanti

- **NIC multiple**: un dispositivo può avere più schede di rete (server, router, ecc.). La tabella `device_interfaces` modella ogni NIC come entità separata (es. eth0, iDRAC, WAN). Ogni NIC può avere più IP (`device_ips`) e una connessione fisica (`device_connections`). Il flag `is_primary` su `device_ips` indica l'IP principale del dispositivo.
- **Switch come dispositivo**: uno switch fisico è un record in `switches` (per gestire le porte) ma NON va duplicato in `devices`. Se si vuole tracciarne IP e seriale, si aggiungono campi direttamente a `switches`.
- **Patch panel**: il collegamento fisico è `device → patch_panel_port → switch_port`. La tabella `device_connections` tiene entrambi i riferimenti opzionali.
- **Import da Excel**: prevedere in futuro un endpoint `POST /api/v1/import/csv`. Non prioritario ora, ma lo schema deve poter accogliere tutti i campi del vecchio Excel.
- **Hostname univoco per sede**: sarebbe utile aggiungere un UNIQUE constraint su `(site_id, hostname)` in `devices` per evitare duplicati.
