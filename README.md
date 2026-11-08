```
              __/\
           __/    \
     ~~~ _/   o    \_
    ~~~ /            \___
   ~~~ /     _          .)
  ~~~ /     / `--------'
 ~~~ /     /
~~~ |     |
```

# Bojack - Runway Manager

`bojack` is a terminal UI for managing your [Runway](https://runway.horse) apps and services.

It is a front end to the `runway` CLI: every view runs a `runway … -o json`
command and renders the result, so it sees exactly what the CLI sees. Actions
that change something (restarts, backups, config changes, plan changes, volume
resizes, domain re-checks) always ask you to confirm first.

## Features

- **Apps:** list with type, plan, region, role and status. The detail view has tabs for
  Info, Processes, Stats, Config, Domains, Sharing and Logs.
- **Stats** (apps and services): one line per metric with its latest value against the plan
  limit and a usage gauge (`━━━━──│───  41%`: filled to the latest value, coloured by the
  warning and critical thresholds, with `│` marking the peak over the interval). Below that,
  a braille line chart of the selected metric (`↑`/`↓`) with time and value axes, scaled
  from 0 to the plan limit, with the thresholds drawn in. Pin a metric (`p`) to chart it
  above the selected one, over the same time range, to compare them. Charts use
  [ntcharts](https://github.com/NimbleMarkets/ntcharts).
- **Services:** details, stats (including database metrics), config, backups and DSN.
  Backups need a paid service plan; dev plans have none.
- **Volumes, Domains, Policies, Regions, Plans, Templates, Registries, SSH Keys, API Keys:**
  simple tables. SSH keys show their `SHA256:` fingerprint, as `ssh-keygen -l` prints it.
  Press `enter` on a template or volume for its details (all versions and the full
  description of a template; whether a volume can be resized, attached or deleted).

## Requirements

- The [`runway` CLI](https://runway.horse) on your `PATH`, logged in (`runway login`).
  Last checked against runway CLI 1.49.0.

## Install / run

Download a binary for your platform from the
[releases page](https://github.com/henrrrik/bojack/releases), or build it
with Go 1.26.8+:

```sh
go install github.com/henrrrik/bojack/cmd/bojack@latest
bojack
```

`bojack --version` prints the version.

Or from a checkout:

```sh
go run ./cmd/bojack
```

## Keys

| Key                 | Action                       |
| ------------------- | ---------------------------- |
| `↑`/`↓`, `j`/`k`    | Move                         |
| `enter`             | Select                       |
| `tab` / `shift+tab` | Next / previous tab (or click a tab) |
| `↑`/`↓`, `j`/`k`    | Stats tab: select the metric to chart |
| `p`                 | Stats tab: pin the selected metric, to chart it above the selected one for comparison; again to unpin |
| `i`                 | Stats tab: cycle the interval (1h / 6h / 24h / 7d / 30d) |
| `f`                 | Stats / Logs tabs: follow (auto-refresh every 15s); again to stop |
| `v`                 | Config / DSN tabs: reveal / hide secrets (hidden by default) |
| `a` / `b` / `s`     | Logs: toggle app / build / system logs |
| `t`                 | Logs: cycle time range (all / 1h / 6h / 24h / 7d) |
| `l`                 | Logs: cycle line limit (100 / 500 / 1000) |
| `/`                 | Lists: filter rows (enter: apply, esc: clear) |
| `r`                 | Refresh                      |
| `o`                 | Open app in browser          |
| `R`                 | Restart app / service (asks to confirm with `y`) |
| `B`                 | Service: create a backup (asks to confirm; not on dev plans) |
| `V`                 | Domains list: re-check verification of the selected domain (asks to confirm) |
| `S` / `D`           | App / service Config tab: set `KEY=value` / remove a key (asks to confirm; secret values are masked as you type) |
| `P`                 | App Info tab: move the app to another plan (asks to confirm) |
| `S`                 | Volume details: grow the volume to a new size in MiB (only for resizable volumes; asks to confirm) |
| `esc`               | Back                         |
| `q`, `ctrl+c`       | Quit                         |

bojack turns on mouse reporting so tabs can be clicked. Most terminals still let you
select text by holding `shift` (`option` in iTerm2 and Terminal.app) while dragging.

## Development

```sh
go test -v -race ./...
golangci-lint fmt
golangci-lint run ./...
```

Linting uses [golangci-lint](https://golangci-lint.run) v2 (config in `.golangci.yml`):
the standard linters (`go vet`, staticcheck, errcheck, …), gofmt `-s`, and gocyclo
with a cyclomatic complexity limit of 10.

### Checking compatibility with a new runway CLI

After updating the CLI (`runway selfupdate`), run:

```sh
go test -tags compat -run TestCLICompat -v ./internal/runway/
```

It calls every read-only command bojack uses against your logged-in account
(first running app and first service) and fails on command or decode errors, or
on fields our types expect that the CLI no longer sends. New CLI fields are
logged. It prints field paths only, never values. CI skips it because it needs
real credentials. If it passes, bump "Last checked against" above.

The CLI is accessed through the `runway.Runner` / `CommandExecutor` interfaces in
`internal/runway`, so tests use fakes instead of calling the real binary. The JSON
fixtures in `internal/runway/types_test.go` mirror real CLI output. If the CLI changes
its output format, update them first.

### Releasing

1. Move the `[Unreleased]` notes in `CHANGELOG.md` under a new version heading and
   merge that in a PR.
2. Tag `main` and push the tag: `git tag v0.2.0 && git push origin v0.2.0`.

The Release workflow runs the tests and then [GoReleaser](https://goreleaser.com),
which builds the binaries (with the version set via `-X main.version`) and publishes a
GitHub release whose notes are that version's `CHANGELOG.md` section.
