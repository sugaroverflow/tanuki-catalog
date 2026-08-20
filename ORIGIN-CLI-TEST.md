# PR tooling test

Throwaway file used to verify pull request creation against this repo.

## Finding

`origin pr create` does **not** work here. Origin refuses it server-side:

```
Origin pull requests are not available for GitHub-mirrored repos;
got mirrorStatus="inbound"
```

`origin repo view` confirms `Mirror status: inbound` — this repo is mirrored
from GitHub, so GitHub stays the forge of record and `gh pr create` is the
path that works. Note that `git push` to the `origin.cursor.com` remote does
proxy through to GitHub, so the branch pushed fine even though the PR did not.

## Scope

Doc-only, so it misses the `registry/**`, `**/*.go`, `cmd|internal/**`, and
`infra/**` path filters. Only the repo-wide `CI` and `Lint` checks should run.

Safe to delete along with this change.
