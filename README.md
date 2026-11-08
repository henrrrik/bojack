# runway-tui

A terminal UI for browsing your [Runway](https://runway.horse) apps and services.

It is a read-only front end to the `runway` CLI: every view runs a `runway … -o json`
command and renders the result, so it sees exactly what the CLI sees.

## Features

- **Apps:** list with type, plan, region, role and status. The detail view has tabs for
  Info, Processes, Stats (CPU, memory, disk and network with usage vs. plan limit and sparklines), Config, Domains,
  Sharing and Logs.
- **Services:** details, stats (including database metrics), config, backups and DSN.
- **Volumes, Domains, Policies, Regions, Plans, Templates, Registries, SSH Keys, API Keys:**
  simple tables. SSH keys show their `SHA256:` fingerprint, as `ssh-keygen -l` prints it.

## Requirements

- Go 1.26+
- The [`runway` CLI](https://runway.horse) on your `PATH`, logged in (`runway login`).
  Last checked against runway CLI 1.49.0.

## Install / run

```sh
go install github.com/henrrrik/runway-tui/cmd/runway-tui@latest
runway-tui
```

Or from a checkout:

```sh
go run ./cmd/runway-tui
```

## Keys

| Key                 | Action                       |
| ------------------- | ---------------------------- |
| `↑`/`↓`, `j`/`k`    | Move                         |
| `enter`             | Select                       |
| `tab` / `shift+tab` | Next / previous tab          |
| `i`                 | Cycle stats interval (1h / 6h / 24h / 7d / 30d) |
| `a` / `b` / `s`     | Logs: toggle app / build / system logs |
| `t`                 | Logs: cycle time range (all / 1h / 6h / 24h / 7d) |
| `l`                 | Logs: cycle line limit (100 / 500 / 1000) |
| `r`                 | Refresh                      |
| `o`                 | Open app in browser          |
| `esc`               | Back                         |
| `q`, `ctrl+c`       | Quit                         |

## Development

```sh
go test -v -race ./...
golangci-lint fmt
golangci-lint run ./...
```

Linting uses [golangci-lint](https://golangci-lint.run) v2 (config in `.golangci.yml`):
the standard linters (`go vet`, staticcheck, errcheck, …), gofmt `-s`, and gocyclo
with a cyclomatic complexity limit of 10.

The CLI is accessed through the `runway.Runner` / `CommandExecutor` interfaces in
`internal/runway`, so tests use fakes instead of calling the real binary. The JSON
fixtures in `internal/runway/types_test.go` mirror real CLI output. If the CLI changes
its output format, update them first.
