# Changelog

All notable changes to bojack. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.0] - 2026-11-08

First release. Checked against runway CLI 1.49.0.

### Added

- Apps: list with type, plan, region, role and status; detail tabs for Info,
  Processes, Stats, Config, Domains, Sharing and Logs.
- Services: details, stats (including database metrics), config, backups and DSN.
- Stats for apps and services: an overview line per metric (latest value
  against the plan limit, plus a usage gauge with the interval's peak marked)
  and a line chart of the selected
  metric with time and value axes and the warning/critical thresholds. Intervals
  from 1h to 30d. Pin a metric (`p`) to chart it above another, over the same
  time range, to compare them.
- Logs: filter by app, build and system logs, time range and line limit.
- Follow mode (`f`) auto-refreshes stats and logs every 15 seconds.
- Lists for volumes, domains, policies, regions, plans, templates, registries,
  SSH keys and API keys, with a `/` filter. Template and volume detail views.
- Secrets in config and DSNs are masked until you press `v`.
- Actions, each behind a confirmation: restart an app or service, create a
  service backup, re-check domain verification, set or remove app and service
  config keys, move an app to another plan and grow a volume.
- Click a tab to switch to it.
- `--version` flag.
- Release binaries for Linux, macOS and Windows.

[Unreleased]: https://github.com/henrrrik/bojack/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/henrrrik/bojack/releases/tag/v0.1.0
