# A8S TUI

A LazyVim-style terminal UI for A8S. It logs in with Keycloak, keeps tokens in memory for the current session, fetches live projects from the backend, and supports single-instance database deployment with live deployment logs.

## Requirements

- Go `1.26.3` or newer
- A terminal with color support
- Browser access for Keycloak login
- A configured A8S backend

## Environment

The TUI is standalone and reads environment values in this order, with later sources overriding earlier ones:

1. `tui/.env`
2. Process environment variables

Copy any needed values from the frontend env into `tui/.env`; the TUI does not require `a8s-frontend` or `a8s-backend` to be present.

Required Keycloak values:

```env
KEYCLOAK_URL=https://keycloak.example.com
KEYCLOAK_REALM=a8s
KEYCLOAK_CLIENT_ID=a8s-frontend
KEYCLOAK_CLIENT_SECRET=replace-with-client-secret
KEYCLOAK_REDIRECT_URL=http://localhost:8250/callback
```

Backend URL values are checked in this order:

```env
BACKEND_API_BASE_URL=
BACKEND_API_URL=
API_URL=
NEXT_PUBLIC_BACKEND_API_BASE_URL=
NEXT_PUBLIC_BACKEND_API_URL=
NEXT_PUBLIC_API_URL=
```

If no backend URL is set, the TUI falls back to:

```env
http://localhost:8080
```

Do not commit real secrets.

## Run

From the TUI directory:

```bash
cd tui
go run .
```

To run tests:

```bash
cd tui
go test ./...
```

## Login Flow

1. Start the TUI with `go run .`.
2. Press `enter` on the launcher, or press `l`.
3. Your browser opens the Keycloak login page.
4. After login, Keycloak redirects to `KEYCLOAK_REDIRECT_URL`.
5. The browser shows a login success page.
6. Return to the terminal; projects load automatically.

Tokens are kept in memory only. Closing the TUI clears the session.

## Project Dashboard

After login, the dashboard shows:

- Left sidebar: project categories and counts
- Main pane: live project list
- Right pane: selected project details
- Bottom statusline: count, refresh time, and key hints

Project data is fetched with the logged-in Keycloak token.

## Database Deployment

Press `d` after login to deploy a single-instance database.

Fields:

- Project name
- Engine
- Database name
- Username
- Password
- Version
- Size

Supported engines in the form:

- PostgreSQL
- MySQL
- MongoDB
- Redis
- Cassandra

After submit, the TUI opens a deployment log screen and polls the backend deployment detail endpoint. Logs are shown from the backend `statusLog`.

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
up/down or j/k  move project selection
left/right      change focused pane
tab             cycle focused pane
/               filter projects
backspace       clear filter
r               refresh live projects
d               deploy database
o               logout
q, esc, ctrl+c  quit
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
b or esc        back to projects
q or ctrl+c     quit
```

## Troubleshooting

If the TUI says required env values are missing, check `tui/.env`.

If login does not return to the TUI, confirm `KEYCLOAK_REDIRECT_URL` uses the same localhost callback configured in Keycloak.

If project loading fails after login, confirm the backend URL is correct and the backend accepts the Keycloak JWT.

If deployment logs do not update, the deployment may not have returned an ID or the backend may not be writing `statusLog` yet.
