# a8s-cli

The official command-line interface and terminal dashboard for Autonomous 8s (A8S). It securely authenticates with Keycloak, connects to your backend, and lets you manage workspace projects and database deployments directly from your terminal.

## Installation

You can install the A8S CLI using any of the following methods.

### NPM

Run without installing:

```bash
npx a8s-cli
```

Or install globally:

```bash
npm install -g a8s-cli
```

Run it with:

```bash
a8s-cli
```

### PNPM

Run without installing:

```bash
pnpm dlx a8s-cli
```

Or install globally:

```bash
pnpm add -g a8s-cli
```

Run it with:

```bash
a8s-cli
```

### Yarn

Run without installing:

```bash
yarn dlx a8s-cli
```

Or install globally:

```bash
yarn global add a8s-cli
```

Run it with:

```bash
a8s-cli
```

### Bun

Run without installing:

```bash
bunx a8s-cli
```

Or install globally:

```bash
bun add -g a8s-cli
```

Run it with:

```bash
a8s-cli
```

The Node package installs a tiny launcher that detects your OS/CPU and runs the bundled native Go binary.

### Homebrew

For macOS and Linux:

```bash
brew tap ITProfessional-Gen01/a8s-cli
brew install a8s-cli
```

Run it with:

```bash
a8s-cli
```

### Go Install

For Go developers, install from the current Go module:

```bash
go install github.com/ITProfessional-Gen01/a8s-cli@latest
```

Run it with:

```bash
a8s-cli
```

If you want the local binary name to be `a8s-cli`, build it explicitly:

```bash
go build -o a8s-cli .
./a8s-cli
```

## Running Locally

From this repository:

```bash
go run main.go
```

Or build and run:

```bash
go build -o a8s-cli .
./a8s-cli
```

If you installed through NPM, PNPM, Yarn, Bun, or Homebrew, use:

```bash
a8s-cli
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

Releases are fully automated via Jenkins, GoReleaser, and NPM. To publish a new version to the world:

1. Ensure your NPM and GitHub tokens are set in your Jenkins Credentials Manager.
2. Push a new Git tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```
Jenkins will automatically build binaries for all platforms, update Homebrew, and publish to NPM!
