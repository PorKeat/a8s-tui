# A8S TUI

A Bubble Tea terminal dashboard for A8S. It logs in with Keycloak, keeps tokens in memory for the current session, fetches live workspace projects from the backend, and supports single-instance database deployment with live deployment logs.

## Features

- Launcher screen with only `Login with Keycloak` and `Quit`
- Smooth dashboard layout with a sidebar, full-content workspace, padding, outlines, and Nerd Font icons
- Workspace section with `Projects` and `Deployment`
- Security section with `Image Scanner`
- Observability section with `Logs` and `Monitoring`
- Settings section with `User` preferences and light/dark mode switching
- Project detail overview with connection profile, backup history, and real database connection data when the backend returns it
- Deployment type picker for `Single database`, `Database cluster`, `Monolithic`, and `Microservices`
- Single database deployment form with engine, version, and size selectors
- Live deployment log screen that polls backend deployment status

The UI is built with:

- Bubble Tea for app state, input, and update loop
- Lip Gloss for colors, borders, padding, and layout
- Nerd Font glyphs for sidebar and project icons
- Bubbles-compatible patterns for terminal components

## Requirements

- Go `1.26.3` or newer
- A terminal with color support
- A Nerd Font installed and selected in your terminal
- Browser access for Keycloak login
- A configured A8S backend

## Run

From the TUI directory:

```bash
cd tui
go run .
```

Run tests:

```bash
cd tui
go test ./...
```

## Login Flow

1. Start the TUI with `go run .`.
2. Select `Login with Keycloak`.
3. Your browser opens the Keycloak login page.
4. After login, Keycloak redirects to `KEYCLOAK_REDIRECT_URL`.
5. Return to the terminal; the TUI loads live projects automatically.

Tokens are kept in memory only. Closing or logging out clears the session.

## Dashboard

After login, the sidebar contains:

```text
Workspace
  Projects
  Deployment

Security
  Image Scanner

Observability
  Logs
  Monitoring

Settings
  User
```

The `Projects` page shows live database deployments, monolith apps, and microservice workspaces returned by the backend. Press `enter` on a project to open its detail page.

## Project Detail

The project detail page shows:

- Overview data: engine, mode, namespace, and updated time
- Connection profile: hostname, port, database, username, engine, version, namespace, and JDBC URL when available
- Backup history placeholder

For database projects, the TUI hydrates connection details by fetching the first deployment ID from `databaseDeploymentIds` through:

```text
GET /api/v1/database-deployments/{deploymentId}
```

`Hostname`, `Port`, and `JDBC URL` are rendered only when real backend data is available. They are not hardcoded.

## Deployment

Open `Deployment` from the sidebar or press `d`.

The deployment page lists:

- Single database
- Database cluster
- Monolithic
- Microservices

Only `Single database` is currently active. Other deployment types are shown as coming soon.

The single database form supports:

- Project name
- Engine
- Database name
- Username
- Password
- Version
- Size

Supported engines:

- PostgreSQL
- MySQL
- MongoDB
- Redis
- Cassandra

After submit, the TUI opens a deployment log screen and polls:

```text
GET /api/v1/database-deployments/{deploymentId}
```

Logs are shown from backend `statusLog`.

## User Settings

Open `User` under Settings to change appearance.

Press `t`, `space`, or `enter` to switch between:

- Dark mode
- Light mode

## Keyboard

Launcher:

```text
up/down or j/k  move
enter           select
l               login
q               quit
```

Dashboard:

```text
up/down or j/k  move selected project, sidebar item, or deployment type
left/right      cycle focus area
tab             cycle focus area
enter           open selected project, sidebar page, or deployment type
/               filter projects
backspace       clear project filter
r               refresh live projects
p               open Projects
d               open Deployment
i               open Image Scanner
g               open Logs
m               open Monitoring
u or s          open User settings
o               logout
esc             close current section back to Projects, or quit from Projects
q or ctrl+c     quit
```

Project detail:

```text
b or esc        back to Projects
q or ctrl+c     quit
```

Deployment type list:

```text
up/down or j/k  move deployment type
enter           open selected deployment type
esc             close Deployment back to Projects
q or ctrl+c     quit
```

Database form:

```text
up/down or j/k  move between fields
left/right      change engine, version, or size
enter           next field, or deploy on the Deploy row
backspace       delete text
esc             cancel
ctrl+c          quit
```

Deployment logs:

```text
up/down or j/k  scroll logs
r               refresh logs now
b or esc        back to Projects and refresh project list
q or ctrl+c     quit
```

User settings:

```text
t, space, enter switch light/dark mode
esc             back to Projects
q or ctrl+c     quit
```

## Troubleshooting

If the TUI says required env values are missing, check `tui/.env`.

If login does not return to the TUI, confirm `KEYCLOAK_REDIRECT_URL` uses the same localhost callback configured in Keycloak.

If projects fail to load after login, confirm the backend URL is correct and the backend accepts the Keycloak JWT.

If Hostname, Port, or JDBC URL do not appear in project detail, confirm the backend deployment detail response includes `serviceHost` and `servicePort`.

If deployment logs do not update, the deployment may not have returned an ID or the backend may not be writing `statusLog` yet.
