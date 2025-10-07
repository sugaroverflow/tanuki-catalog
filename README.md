# Tanuki Service Catalog

`tanuki` is the service catalog CLI for the Tanuki Services platform. The
`registry/` directory is the source of truth: one YAML manifest per service,
validated against `catalog.schema.json` and compiled into `catalog.json` by CI.

## Quick start

```sh
go build -o tanuki ./cmd/tanuki
./tanuki list
```

No Python required — the committed `catalog.json` (and the registry manifests
themselves) are enough to run every read command.

## Commands

| Command | What it does |
|---|---|
| `tanuki list` | List all registered services |
| `tanuki status <name>` | Show health, version, owner, last deploy |
| `tanuki owners <name>` | Show owner and on-call info |
| `tanuki search --team <team>` | Filter services by team |
| `tanuki validate` | Validate registry manifests against the schema |
| `tanuki version` | Print the CLI version |

The catalog is read from `TANUKI_CATALOG_URL` (set `TANUKI_CATALOG_KEY` for
authenticated endpoints), a local `catalog.json` / `dist/catalog.json`, or
built directly from `registry/` manifests as a fallback.

## Adding a service

1. Add `registry/<service-name>.yml` — the manifest `name` must match the
   filename. See `catalog.schema.json` for required fields.
2. Run the validator:
   ```sh
   python3 -m venv .venv && .venv/bin/pip install -r requirements.txt
   .venv/bin/python3 scripts/build_catalog.py --validate
   ```
3. Open a pull request. CI validates changed manifests and rebuilds the
   published catalog on merge.

## Repository layout

| Path | Purpose |
|---|---|
| `cmd/tanuki/` | CLI entry point |
| `internal/` | Catalog loading, formatting, schema bridge |
| `registry/` | Service manifests (source of truth) |
| `scripts/build_catalog.py` | Validates manifests, builds `catalog.json` |
| `site/` | Static catalog viewer (published via Pages) |
| `infra/` | Terraform for the catalog snapshot bucket |
| `.github/workflows/` | CI/CD workflows |

## License

MIT — see [LICENSE](LICENSE).
