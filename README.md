# runway-tui

A terminal UI for browsing your [Runway](https://runway.horse) apps and services.

It is a read-only front end to the `runway` CLI: every view runs a `runway … -o json`
command and renders the result, so it sees exactly what the CLI sees.

## Features

- **Apps:** list with type, plan, region, role and status. The detail view has tabs for
  Info, Processes, Stats (CPU, RAM, disk and network sparklines), Config, Domains,
  Sharing and Logs.
- **Services:** details, DSN, config and backups.
- **Volumes, Domains, Policies, Regions, Plans:** simple tables.

## Requirements

- Go 1.26+
- The [`runway` CLI](https://runway.horse) on your `PATH`, logged in (`runway login`).
  Last checked against runway CLI 1.47.1.

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
| `r`                 | Refresh                      |
| `o`                 | Open app in browser          |
| `esc`               | Back                         |
| `q`, `ctrl+c`       | Quit                         |

## Development

```sh
go test -v -race ./...
gofmt -s -w .
go vet ./...
gocyclo -over 10 .
```

The CLI is accessed through the `runway.Runner` / `CommandExecutor` interfaces in
`internal/runway`, so tests use fakes instead of calling the real binary. The JSON
fixtures in `internal/runway/types_test.go` mirror real CLI output. If the CLI changes
its output format, update them first.
