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
- Autenticazione JWT (username/password) con middleware `AuthRequired`
- Audit trail: ogni operazione CRUD viene loggata in `audit_logs`
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
│   ├── main.go                 # Entry point: subcommands (version, update) + bubbletea TUI
│   ├── go.mod                  # module github.com/guerrieroriccardo/CIPTr/cli
│   └── internal/
│       ├── version/            # Version/Commit/Date vars (injected via ldflags)
│       │   └── version.go
│       ├── selfupdate/         # Self-update from GitHub Releases
│       │   └── selfupdate.go
│       ├── apiclient/          # Client HTTP per la REST API
│       │   ├── client.go       # HTTP helpers + envelope parsing
│       │   └── clients.go      # Metodi per ogni risorsa (uno per file)
│       └── tui/                # Componenti bubbletea
│           ├── app.go          # Root model + dispatching
│           ├── nav.go          # Stack di navigazione + breadcrumb
│           ├── styles.go       # Stili lipgloss
│           ├── menu.go         # Menu principale (mostra versione nel footer)
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
├── .goreleaser.yml             # GoReleaser: cross-platform CLI builds
├── .github/workflows/
│   └── release.yml             # CI: build + publish on version tags
├── .gitignore
└── CLAUDE.md                   # Questo file
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

manufacturers (lookup: HP, Cisco, Dell, ...)
categories (lookup: Server, PC, Switch, Printer, ...)
suppliers (company info: name, address, phone, email)

device_models (hardware catalog, FK → manufacturers, FK → categories)
  └── devices (deployed devices, FK → categories, FK → suppliers)
        └── device_interfaces (NICs: eth0, iDRAC, WAN, LAN1, ...)
              └── device_ips (IP per NIC, linked to a VLAN)
              └── device_connections (physical link: NIC → switch_port ↔ patch_panel_port)
```

### SQL — schema.sql

Vedi `backend/db/schema.sql` per lo schema completo (PostgreSQL).

Tabelle principali e colonne chiave:

| Tabella | Colonne chiave | Note |
|---------|---------------|------|
| `manufacturers` | `id`, `name` (UNIQUE) | Lookup: HP, Cisco, Dell... |
| `categories` | `id`, `name` (UNIQUE) | Lookup: Server, PC, Switch... |
| `suppliers` | `id`, `name`, `address`, `phone`, `email` | Azienda fornitrice |
| `clients` | `id`, `name`, `short_code` | Cliente MSP |
| `sites` | `id`, `client_id` (FK), `name`, `address` | Sede fisica |
| `locations` | `id`, `site_id` (FK), `name`, `floor` | Stanza/piano |
| `address_blocks` | `id`, `site_id` (FK), `network` (CIDR) | Blocco IP assegnato |
| `vlans` | `id`, `site_id` (FK), `vlan_id`, `subnet` (CIDR) | Sottorete |
| `device_models` | `id`, `manufacturer_id` (FK), `model_name`, `category_id` (FK) | Catalogo HW |
| `devices` | `id`, `site_id` (FK), `hostname`, `category_id` (FK), `supplier_id` (FK) | Dispositivo deployato |
| `device_interfaces` | `id`, `device_id` (FK), `name`, `mac_address` (MACADDR) | NIC |
| `device_ips` | `id`, `interface_id` (FK), `ip_address` (INET) | IP per NIC |
| `device_connections` | `id`, `interface_id` (FK), `switch_port_id`, `patch_panel_port_id` | Collegamento fisico |
| `switches` | `id`, `site_id` (FK), `name`, `ip_address` (INET) | Switch di rete |
| `switch_ports` | `id`, `switch_id` (FK), `port_number` | Porta dello switch |
| `patch_panels` | `id`, `site_id` (FK), `name` | Patch panel |
| `patch_panel_ports` | `id`, `patch_panel_id` (FK), `port_number` | Porta del patch panel |

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
| GET | `/devices` | Lista dispositivi (filtri: site_id, status, category_id, search) |
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

### Manufacturers (Lookup)
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/manufacturers` | Lista produttori |
| POST | `/manufacturers` | Crea produttore |
| GET | `/manufacturers/:id` | Dettaglio produttore |
| PUT | `/manufacturers/:id` | Aggiorna produttore |
| DELETE | `/manufacturers/:id` | Elimina produttore |

### Categories (Lookup)
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/categories` | Lista categorie |
| POST | `/categories` | Crea categoria |
| GET | `/categories/:id` | Dettaglio categoria |
| PUT | `/categories/:id` | Aggiorna categoria |
| DELETE | `/categories/:id` | Elimina categoria |

### Suppliers
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/suppliers` | Lista fornitori |
| POST | `/suppliers` | Crea fornitore |
| GET | `/suppliers/:id` | Dettaglio fornitore |
| PUT | `/suppliers/:id` | Aggiorna fornitore |
| DELETE | `/suppliers/:id` | Elimina fornitore |

### Device Models (Catalogo)
| Metodo | Path | Descrizione |
|--------|------|-------------|
| GET | `/device-models` | Lista modelli (filtri: category_id, manufacturer_id) |
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
Tutte le 17 risorse CRUD implementate:
clients, sites, locations, address_blocks, vlans,
manufacturers, categories, suppliers,
device_models, devices, device_interfaces, device_ips, device_connections,
switches, switch_ports, patch_panels, patch_panel_ports

### Fase 2b — CLI Releases & Self-Update ✅
1. ✅ Version embedding via `-ldflags` (`cli/internal/version/`)
2. ✅ Subcommands: `ciptr-cli version`, `ciptr-cli update` (before TUI starts)
3. ✅ Self-update from GitHub Releases (`go-selfupdate` → `guerrieroriccardo/CIPTr`)
4. ✅ GoReleaser config (`.goreleaser.yml`): builds `linux/amd64` + `windows/amd64`
5. ✅ GitHub Actions CI (`.github/workflows/release.yml`): triggered on `v*` tags
6. ✅ Version shown in TUI menu footer

**Release workflow:** `git tag v1.0.0 && git push --tags` → GitHub Actions → GoReleaser → GitHub Release with binaries → `ciptr-cli update` picks it up.

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
- Questo documento (`CLAUDE.md`) rimane in italiano perché è una specifica per il team.
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
- Subcommands (`version`, `update`) vengono gestiti prima di avviare la TUI
- Versione iniettata a build time via `-ldflags` (vedi `.goreleaser.yml`)
- Self-update scarica da GitHub Releases (`guerrieroriccardo/CIPTr`)
- Target: `linux/amd64`, `windows/amd64`
- Componenti generici (table, form) guidati da un registro di definizioni (`resource/registry.go`)
- Navigazione stack-based: `PushScreenMsg` / `PopScreenMsg` / Esc per tornare indietro
- **Tutte le tabelle devono essere filtrabili**: pressione `/` attiva un campo di ricerca che filtra le righe per qualsiasi colonna (case-insensitive). Enter conferma il filtro, Esc lo cancella. Questo vale sia per il menu principale (built-in di `bubbles/list`) che per le tabelle risorse (`ResourceTable` con `textinput`)
- Login screen con persistenza token in `~/.config/ciptr/token`; redirect automatico al login su 401

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

### Rilasci CLI
- Quando si aggiunge una feature significativa o un bug fix alla CLI, **creare un nuovo tag di versione** per triggerare una release:
  `git tag v<MAJOR>.<MINOR>.<PATCH> && git push --tags`
- Seguire semver: MAJOR per breaking changes, MINOR per nuove feature, PATCH per bug fix
- Il CI (`.github/workflows/release.yml`) builda automaticamente i binari per linux/amd64 e windows/amd64
- Gli utenti ricevono l'aggiornamento tramite `ciptr-cli update`

### Database
- PostgreSQL 18 — FK enforcement è attivo di default
- Nessun soft delete: eliminazione reale con conferma nel frontend/CLI
- Schema in `backend/db/schema.sql`

---

## 8. Valori Enumerati

### devices.status
`active`, `inactive`, `reserved`, `decommissioned`

### categories, manufacturers, suppliers (tabelle lookup)
Sono tabelle dinamiche gestite via API CRUD — l'utente può aggiungere, modificare o eliminare valori liberamente senza toccare il codice.
Esempi di categorie iniziali: `PC`, `Laptop`, `Server`, `Printer`, `Switch`, `Router`, `AP`, `NAS`, `Camera`, `Phone`, `UPS`, `Other` — ma qualsiasi valore è valido (es. `Firewall`, `Storage`, `Tablet`...).

---

## 9. Note Importanti

- **NIC multiple**: un dispositivo può avere più schede di rete (server, router, ecc.). La tabella `device_interfaces` modella ogni NIC come entità separata (es. eth0, iDRAC, WAN). Ogni NIC può avere più IP (`device_ips`) e una connessione fisica (`device_connections`). Il flag `is_primary` su `device_ips` indica l'IP principale del dispositivo.
- **Switch come dispositivo**: uno switch fisico è un record in `switches` (per gestire le porte) ma NON va duplicato in `devices`. Se si vuole tracciarne IP e seriale, si aggiungono campi direttamente a `switches`.
- **Patch panel**: il collegamento fisico è `device → patch_panel_port → switch_port`. La tabella `device_connections` tiene entrambi i riferimenti opzionali.
- **Import da Excel**: prevedere in futuro un endpoint `POST /api/v1/import/csv`. Non prioritario ora, ma lo schema deve poter accogliere tutti i campi del vecchio Excel.
- **Hostname univoco per sede**: sarebbe utile aggiungere un UNIQUE constraint su `(site_id, hostname)` in `devices` per evitare duplicati.
