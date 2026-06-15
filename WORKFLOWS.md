# CI/CD workflows

How the Tanuki catalog repo builds, tests, and ships. If you change a
workflow, update this file — it is the only place the whole picture exists.

## The core loop

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | every push + PR | Validate registry, test matrix (Go 1.24/1.25 × Linux/macOS/Windows, minus 1.24-on-Windows — PLAT-1861), cross-compile artifacts |
| `lint.yml` | every push + PR | golangci-lint + pre-commit hooks |
| `integration.yml` | app code changes | Remote-catalog path against a WireMock service container |
| `codeql.yml` | Go changes + weekly | Static analysis (`build-mode: manual` — CodeQL needs the full-compile incantation, do not remove `go clean -cache`) |
| `registry-lint.yml` | registry changes | Validates changed manifests, classifies affected teams, comments on the PR |

## Shipping

| Workflow | Trigger | What it does |
|---|---|---|
| `publish.yml` | merge to main | catalog → binary → staging → smoke-test → **production (manual approval)**. Smoke test asserts the CLI row count matches the built catalog's service count |
| `container-publish.yml` | app changes on main | Multi-arch image to GHCR (`type=sha` + `latest`) |
| `deploy-catalog-pages.yml` | registry/site changes | Publishes the static catalog viewer to Pages |
| `release.yml` | `v*` tags | Cross-compile 5 platforms, checksums, SBOM, GPG-sign (skipped if no signing key), GitHub Release with generated notes |
| `reusable-go-build.yml` | called by publish + release | The one true Go build: platforms/version in, artifact + version out |

## Per-service deploys

One workflow per service (`deploy-<service>.yml`): bumping
`registry/<service>.yml` on main republishes the manifest and force-redeploys
that service. Most target AWS (role `tanuki-deploy-<team>`); `data-pipeline`
is on GCP and `reporting-service` is still on the old Azure subscription
(DATA-311). Rollbacks are manual dispatch with a task-definition revision
(`rollback-*.yml`); `canary-payments-api.yml` shifts a traffic percentage.

These pre-date `reusable-go-build.yml` and should eventually be folded into a
reusable workflow (PLAT-1742, PAY-988).

## Scheduled

| Workflow | Schedule | What it does |
|---|---|---|
| `nightly-catalog-validate.yml` | 02:00 daily | Registry validation (with retry), drift check against the committed catalog, benchmark run |
| `catalog-refresh.yml` | Mondays 06:00 | Rebuilds `catalog.json`, opens a PR when it drifts |
| `notify-oncall.yml` | on nightly failure | Pages #platform-oncall via Slack |
| `smoke-payments.yml` / `smoke-platform.yml` | registry changes | Team-scoped `tanuki search` / `status` checks |

## Infra

`infra-plan.yml` runs `terraform fmt`/`validate` on `infra/**` PRs; the
cloud-touching steps only run when the `DEPLOY_ENABLED` repo variable is set.

## Composite actions

| Action | Used by | Notes |
|---|---|---|
| `.github/actions/setup-tanuki-env` | test/build/smoke jobs | Go toolchain only; module caching is the caller's job (see `ci.yml`'s cache step, keyed on `go.sum`) |
| `.github/actions/catalog-tools` | validate/publish/nightly/refresh/pages | Python + pip cache + catalog build; outputs `catalog-path` and `service-count` |

## Conventions

- Environments: `staging` (auto), `production` (requires approval), plus
  per-team `platform`/`payments`/`experience`/`data` for deploys.
- `ci.yml` cancels superseded runs per ref; `publish.yml` **queues** instead
  (`cancel-in-progress: false`) — never cancel a deploy in flight.
- Secrets: `SLACK_WEBHOOK_URL`, `GPG_PRIVATE_KEY`/`GPG_PASSPHRASE`,
  `AZURE_*`. Cloud deploy steps are additionally gated on the
  `DEPLOY_ENABLED` repo variable.
- Artifact retention: 7 days for CI binaries, 14 for the published catalog,
  3 for nightly drift reports.
