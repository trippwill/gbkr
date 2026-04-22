# Repository Guidelines

## What This Module Is

`gbkr` — Permission-gated Go client library for the Interactive Brokers REST API. Public repo with branch protection; changes go through PRs.

## Quick Start

Run `mise tasks` to see available build/test/lint tasks. Prefer `mise` tasks over raw Go commands.

## Development Context

For full architecture, API patterns, and conventions, use the **`gbkr-dev`** skill. It covers:
- Two-tier client model (gateway vs brokerage session)
- Operation event sourcing (ADR-007)
- Package roles and capability separation (ADR-008)
- Go code style, error handling, and naming conventions
- Build/test/lint commands and CI pipeline

Additional skills: **`testing-guide`** for test conventions, **`workspace-git`** for commit/push workflows.

## Invariants

- **Green-field**: No module version declared. Prefer better design over backward compatibility.
- **Financial types**: All financial fields use `num.Num` / `num.NullNum` — never `float64`. See ADR-014.
- Run `mise run precommit` before commits.
- Linting uses `golangci-lint` v2; config in `.golangci.yml`.
