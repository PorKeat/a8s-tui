# a8s-cli

The official command-line interface and terminal dashboard for Autonomous 8s (A8S). It securely authenticates with Keycloak, connects to your backend, and lets you manage workspace projects and database deployments directly from your terminal.

## Installation

You can install `a8s-cli` using any of the following methods:

### 1. NPM (Node Package Manager)
The fastest way for web developers:
```bash
npm install -g a8s-cli
```
*(This automatically detects your OS and downloads the native Go binary—it does not run slow JavaScript!)*

### 2. Homebrew (macOS / Linux)
The standard for Mac/Linux developers:
```bash
brew tap ITProfessional-Gen01/a8s-cli
brew install a8s-cli
```

### 3. Go Install
For Go developers:
```bash
go install github.com/ITProfessional-Gen01/a8s-cli@latest
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
- **Projects Dashboard**: View live database deployments, monoliths, and microservice workspaces.
- **Project Details**: Get connection profiles, hostnames, ports, and JDBC URLs instantly.
- **Database Deployments**: Create single-instance databases (PostgreSQL, MySQL, MongoDB, Redis, Cassandra) with version and size selectors.
- **Live Logs**: Watch real-time deployment logs stream directly to your terminal.
- **Customizable UI**: Fully integrated light/dark mode and support for icon-less environments (set `A8S_NO_ICONS=true`).

## Keyboard Shortcuts

The UI is highly responsive and designed for power users. 

**Global Navigation:**
- `Up/Down` or `j/k`: Move cursor
- `Left/Right` or `Tab`: Switch focus between Sidebar and Main Content
- `Enter`: Select or open
- `q` or `ctrl+c`: Quit
- `esc`: Go back

**Shortcuts (When Ready):**
- `p`: Open Projects
- `d`: Open Deployments
- `i`: Open Image Scanner
- `g`: Open Logs
- `m`: Open Monitoring
- `u` or `s`: Open User Settings
- `r`: Refresh data
- `/`: Filter projects
- `o`: Logout

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

Releases are fully automated via GitHub Actions, GoReleaser, and NPM. To publish a new version to the world:

1. Ensure your NPM token is set in GitHub Secrets as `NPM_TOKEN`.
2. Push a new Git tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```
GitHub Actions will automatically build binaries for all platforms, update Homebrew, and publish to NPM!
