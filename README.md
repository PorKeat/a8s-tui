# a8s-tui

The official command-line interface and terminal dashboard for Autonomous 8s (A8S). It securely authenticates with Keycloak, connects to your backend, and lets you manage workspace projects and database deployments directly from your terminal.

## Installation

You can install the A8S CLI using any of the following methods.

### NPM

Run without installing:

```bash
npx a8s-tui
```

Or install globally:

```bash
npm install -g a8s-tui
```

Run it with:

```bash
a8s-tui
```

### PNPM

Run without installing:

```bash
pnpm dlx a8s-tui
```

Or install globally:

```bash
pnpm add -g a8s-tui
```

Run it with:

```bash
a8s-tui
```

### Yarn

Run without installing:

```bash
yarn dlx a8s-tui
```

Or install globally:

```bash
yarn global add a8s-tui
```

Run it with:

```bash
a8s-tui
```

### Bun

Run without installing:

```bash
bunx a8s-tui
```

Or install globally:

```bash
bun add -g a8s-tui
```

Run it with:

```bash
a8s-tui
```

The Node package installs a tiny launcher that detects your OS/CPU and runs the bundled native Go binary.

### Homebrew

For macOS and Linux:

```bash
brew tap porkeat/a8s-tui https://github.com/PorKeat/a8s-tui
brew install porkeat/a8s-tui/a8s-tui
```

Run it with:

```bash
a8s-tui
```

### Go Install

For Go developers, install from the current Go module:

```bash
go install github.com/PorKeat/a8s-tui@latest
```

Run it with:

```bash
a8s-tui
```

If you want the local binary name to be `a8s-tui`, build it explicitly:

```bash
go build -o a8s-tui .
./a8s-tui
```

## Running Locally

From this repository:

```bash
go run main.go
```

Or build and run:

```bash
go build -o a8s-tui .
./a8s-tui
```

If you installed through NPM, PNPM, Yarn, Bun, or Homebrew, use:

```bash
a8s-tui
```

## Features

- **Authentication**: Seamless Keycloak integration with automatic token refreshing.
- **Projects Dashboard**: View live database deployments, monoliths, and microservice workspaces. Monolith details can run a browser route check and show pass/fail results.
- **Project Details**: Get connection profiles, hostnames, ports, and JDBC URLs instantly.
- **Database Deployments**: Create single-instance databases and database clusters (PostgreSQL, MySQL, MongoDB, Redis, Cassandra) with version and size selectors.
- **Application Deployments**: Deploy monolithic apps or scan mono-repo and multi-repo GitHub sources into a microservice workspace.
- **Image Scanner**: Scan deployed Harbor images, external registry references, or Git repository builds; follow scan progress and review Trivy findings and reports.
- **Logs**: Inspect workspace Kubernetes pods and load recent runtime log output directly in the terminal.
- **Monitoring**: View namespace health, resource usage, pod status, and per-project metrics from the backend monitoring API.
- **Live Deployment Logs**: Watch deployment logs stream directly to your terminal while new workloads are created.
- **Customizable UI**: Fully integrated themes (Dark, Light, Orange, Green, Ocean, Rose) and support for icon-less environments (set `A8S_NO_ICONS=true`).
- **Light by default**: The launcher, dashboard, forms, logs, dialogs, status colors, and monitoring charts all start with the complete light palette.

## Keyboard Shortcuts

The UI is highly responsive and designed for power users. 

**Global Navigation:**
- `Up/Down` or `j/k`: Move cursor
- `Tab`: Switch focus between Sidebar and Main Content
- `Enter`: Select or open
- `q` or `ctrl+c`: Quit
- `esc`: Go back

**Shortcuts (When Ready):**
- `p`: Select Projects, then `Enter` opens it
- `d`: Select Deployments, then `Enter` opens it
- `i`: Select Image Scanner, then `Enter` opens it
- `g`: Select Logs, then `Enter` opens it
- `m`: Select Monitoring, then `Enter` opens it
- `u` or `s`: Select User Settings, then `Enter` opens it
- `r`: Refresh data
- `/`: Filter projects
- `o`: Logout

**User Settings:**
- `Enter`, `Space`, or `t`: Cycle through available themes

**Project Details:**
- Open a deployed monolith, select `Check routes`, then press `Enter` to run the backend browser route check
- `Left/Right`, `Up/Down`, or `Tab`: Switch between project actions

**Microservice Deployment:**
- Select `Mono Repo` to detect multiple services from one repository.
- Select `Multi Repo` to scan and merge services from several repositories.
- Enter a GitHub remote, select `Scan repository`, and review the detected services before deploying.
- Select `Env service` with `Left/Right`, then press `Enter` on `.env file` to browse for and import a local environment file into that service.
- Imported environment values stay in memory for the deployment session. The TUI displays only variable counts and marks secret-like names as secrets.
- Open `Relationships` after detection to choose a source service, target service, and relationship type. The TUI manages both `dependsOn` and generated relationship environment variables before deployment.
- Relationship values use the target service name, matching the web canvas flow; the platform generates the final in-cluster runtime URL.
- The TUI verifies every scanned Git remote again immediately before deployment.
- The TUI supports public GitHub repositories only and never reads or sends a GitHub token.

**Image Scanner:**
- `i`, then `Enter`: Open Image Scanner
- `Left/Right`: Switch between Harbor, External, Git, and History
- `Up/Down` or `j/k`: Move through images, scan history, or source fields
- `Enter`: Advance through source fields, start a scan, or open a history result
- `Space`: Toggle private registry/repository access while its field is selected
- Paste external registry and Git repository values directly into the selected field
- Harbor scans select a deployed image; External scans pull an image reference; Git scans clone, build, and scan the resulting image
- Completed scans show severity counts, findings, and a Trivy JSON report preview
- `n`: Return to Harbor and choose another source after viewing a result
- `x`: Force a fresh rescan while viewing a source result
- `r`: Refresh images and scan history

**Observability:**
- `g`, then `Enter`: Open Logs
- `m`, then `Enter`: Open Monitoring
- `Up/Down` or `j/k`: Move between pods or project metrics
- `Enter`: Reload logs for the selected pod
- `r`: Refresh the current observability view

## Development

To build the binary locally:
```bash
go build .
```

To run the test suite:
```bash
go test ./...
```

To quickly commit and push your code:
```bash
./auto_push.sh "your commit message"
```

## Publishing a Release

Releases are fully automated via Jenkins, GoReleaser, and NPM. To publish a new version to the world:

1. Ensure your NPM and GitHub tokens are set in your Jenkins Credentials Manager.
2. Push a new Git tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```
Jenkins will automatically build binaries for all platforms, update Homebrew, and publish to NPM!
