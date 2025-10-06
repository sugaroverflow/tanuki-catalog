#!/usr/bin/env python3
"""
Build catalog.json from registry/*.yml manifests.
Validates each manifest against catalog.schema.json.
Run from repo root: python scripts/build_catalog.py [--validate] [--output PATH]
"""
import argparse
import json
import sys
from pathlib import Path

import yaml
from jsonschema import Draft7Validator, FormatChecker

REPO_ROOT = Path(__file__).resolve().parent.parent
REGISTRY_DIR = REPO_ROOT / "registry"
SCHEMA_PATH = REPO_ROOT / "catalog.schema.json"
DEFAULT_OUTPUT = REPO_ROOT / "dist" / "catalog.json"


def load_schema():
    with open(SCHEMA_PATH, encoding="utf-8") as f:
        return json.load(f)


def load_and_validate_manifests(schema):
    validator = Draft7Validator(schema, format_checker=FormatChecker())
    catalog = []
    errors = []
    seen = {}
    yaml_files = sorted(REGISTRY_DIR.glob("*.yml")) + sorted(REGISTRY_DIR.glob("*.yaml"))
    if not yaml_files:
        print("No .yml/.yaml files in registry/", file=sys.stderr)
        sys.exit(1)

    for path in yaml_files:
        try:
            with open(path, encoding="utf-8") as f:
                data = yaml.safe_load(f)
        except yaml.YAMLError as exc:
            first_line = str(exc).splitlines()[0] if str(exc) else exc.__class__.__name__
            errors.append(f"{path.name}: invalid YAML: {first_line}")
            continue
        if data is None:
            errors.append(f"{path.name}: empty manifest")
            continue

        schema_errors = sorted(err.message for err in validator.iter_errors(data))
        if schema_errors:
            errors.extend(f"{path.name}: {msg}" for msg in schema_errors)
            continue

        name = data["name"]
        if path.stem != name:
            errors.append(f"{path.name}: manifest name {name!r} does not match filename")
            continue
        if name in seen:
            errors.append(f"{path.name}: duplicate service name {name!r} (also defined in {seen[name]})")
            continue
        seen[name] = path.name
        catalog.append(data)

    if errors:
        for message in errors:
            print(message, file=sys.stderr)
        print(f"{len(errors)} error(s) across {len(yaml_files)} manifest(s)", file=sys.stderr)
        sys.exit(1)
    return catalog


def main():
    parser = argparse.ArgumentParser(description="Build catalog from registry YAMLs")
    parser.add_argument("--validate", action="store_true", help="Only validate; do not write catalog")
    parser.add_argument("--output", "-o", type=Path, default=DEFAULT_OUTPUT, help="Output path for catalog.json")
    args = parser.parse_args()

    if not REGISTRY_DIR.is_dir():
        print(f"Registry directory not found: {REGISTRY_DIR}", file=sys.stderr)
        sys.exit(1)
    if not SCHEMA_PATH.is_file():
        print(f"Schema not found: {SCHEMA_PATH}", file=sys.stderr)
        sys.exit(1)

    schema = load_schema()
    catalog = load_and_validate_manifests(schema)

    if args.validate:
        print(f"Validation passed ({len(catalog)} manifests).")
        return

    args.output.parent.mkdir(parents=True, exist_ok=True)
    # Deterministic output: sorted files (above), sorted keys, trailing newline.
    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(catalog, f, indent=2, sort_keys=True)
        f.write("\n")
    print(f"Wrote {len(catalog)} services to {args.output}")


if __name__ == "__main__":
    main()
