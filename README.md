# CIPTr

Do you work for an MSP and can't stand yet another spreadsheet to enter your devices information into?
Or are you a solo IT admin and you randomly choose a hostname every time and hope it's not already taken?

**C**lient **IP** **Tr**acker is here to help you better manage your infrastructure and focus on what's really important.

> [!NOTE]
> This project is, at the moment of writing, purely vibecoded. It is used by a real organization to track hundreds of networks and thousands of devices, but bugs are expected and can be reported.

> [!CAUTION]
> This project is under heavy development and use in production is not recommended. If you really want to deploy this project, make sure your backups are working as intended and are restorable.

## Features

CIPTr gives you a fully featured terminal UI experience to manage your infrastructure.
Using hierarchical or absolute navigation, you can easily browse all your devices without losing track of what matters.

### Views

Navigate your infrastructure the way you prefer — drill down from client to site to device, or jump straight to any resource.

![Hierarchical](./tapes/browse.gif)
![Absolute](./tapes/sections.gif)

### Multi-user with audit trail

Role-based access control lets you decide who can do what, while every change is automatically logged for full traceability.

![Authentication](./tapes/authentication.gif)

### Instant filtering

Every table supports real-time filtering across all fields — just press `/` and start typing.

![Filter](./tapes/filter.gif)

### TUI-native forms

No clunky buttons or slow web pages (yet!). Create and edit any entity directly from the terminal.

![Form](./tapes/form.gif)

### Self-updating

CIPTr can update itself with a single command — no need to re-download anything.

![version](./tapes/version.gif)

## Installation

### CLI (Linux / Windows)

```sh
curl -sSfL https://raw.githubusercontent.com/guerrieroriccardo/CIPTr/main/scripts/install.sh | sh
```

Or download the latest binary from [GitHub Releases](https://github.com/guerrieroriccardo/CIPTr/releases).

### Server (Docker Compose)

Make sure you have [Docker](https://docs.docker.com/get-docker/) installed, then:

```sh
curl -O https://raw.githubusercontent.com/guerrieroriccardo/CIPTr/main/compose.yml
docker compose up -d
```

The backend API will be available at `http://localhost:8080`.
