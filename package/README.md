# a8s-cli

The official command-line interface and terminal dashboard for Autonomous 8s (A8S). It securely authenticates with Keycloak, connects to your backend, and lets you manage workspace projects and database deployments directly from your terminal.

## Installation

You can install `a8s-cli` using any of the following methods:

### 1. NPM (Node Package Manager)
The fastest way for web developers:
```bash
npm install -g a8s-tui
```
*(This automatically detects your OS and downloads the native Go binary—it does not run slow JavaScript!)*

### 2. Homebrew (macOS / Linux)
The standard for Mac/Linux developers:
```bash
brew tap PorKeat/a8s-tui
brew install a8s-cli
```

### 3. Go Install
For Go developers:
```bash
go install github.com/PorKeat/a8s-tui@latest
```

## Running the CLI

If you installed it globally using the commands above, just run:
```bash
a8s-cli
```

If you are developing locally, run:
```bash
go run main.go
```

## Features

- **Authentication**: Seamless Keycloak integration with automatic token refreshing.
- **Projects Dashboard**: View live database deployments, monoliths, and microservice workspaces. Monolith details can run a browser route check and show pass/fail results.
- **Project Details**: Get connection profiles, hostnames, ports, and JDBC URLs instantly.
- **Database Deployments**: Create single-instance databases and database clusters (PostgreSQL, MySQL, MongoDB, Redis, Cassandra) with version and size selectors.
- **Application Deployments**: Deploy monolithic apps or a one-service microservices workspace from a Git remote.
- **Image Scanner**: Load deployed container images, start Trivy scans, poll scan status, and review severity findings. The API client also supports Git-repository scan targets.
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

**Observability:**
- `g`, then `Enter`: Open Logs
- `m`, then `Enter`: Open Monitoring
- `Up/Down` or `j/k`: Move between pods or project metrics
- `Enter`: Reload logs for the selected pod
- `r`: Refresh the current observability view

**Project Details:**
- Open a deployed monolith, select `Check routes`, then press `Enter` to run the backend browser route check
- `Left/Right`, `Up/Down`, or `Tab`: Switch between project actions

**Image Scanner:**
- `i`, then `Enter`: Open Image Scanner
- `Left/Right`: Switch Scan and History
- `Up/Down` or `j/k`: Move selected image or scan
- `Enter`: Scan the selected image or open the selected history result with its Trivy report preview
- `r`: Refresh images and scan history

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
