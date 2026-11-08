# Build & Test
- `go test -v -race ./...` — run tests (matches CI)
- `golangci-lint fmt` — run before committing (gofmt -s)
- `golangci-lint run ./...` — run before committing (go vet, staticcheck, errcheck, gocyclo over 10, etc.; config in `.golangci.yml`)

## Workflow

- Use Red/Green TDD
- Create a PR for all changes — do not push directly to main.
- Use interfaces for test doubles — keep this pattern
