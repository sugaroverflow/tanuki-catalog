package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gitlab.com/gitlab-da/projects/tanuki-services/tanuki-catalog/internal/catalog"
	"gitlab.com/gitlab-da/projects/tanuki-services/tanuki-catalog/internal/formatter"
	"gitlab.com/gitlab-da/projects/tanuki-services/tanuki-catalog/internal/schema"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	cmd, rest := args[0], args[1:]

	switch cmd {
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	case "version", "--version":
		fmt.Fprintf(stdout, "tanuki %s\n", version)
		return 0
	case "list":
		if len(rest) > 0 {
			fmt.Fprintf(stderr, "list takes no arguments (got %q)\n", strings.Join(rest, " "))
			return 2
		}
		return runList(stdout, stderr)
	case "status":
		name, code := parseNameArg("status", rest, stderr)
		if code != 0 {
			return code
		}
		return runStatus(stdout, stderr, name)
	case "owners":
		name, code := parseNameArg("owners", rest, stderr)
		if code != 0 {
			return code
		}
		return runOwners(stdout, stderr, name)
	case "search":
		team, code := parseSearchArgs(rest, stderr)
		if code != 0 {
			return code
		}
		return runSearch(stdout, stderr, team)
	case "validate":
		return runValidate(stdout, stderr)
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n", cmd)
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `tanuki - Service catalog CLI

Usage:
  tanuki list                    List all registered services
  tanuki status <name>           Show health, version, owner, last deploy
  tanuki owners <name>           Show owner and on-call info
  tanuki search --team <team>    Filter services by team
  tanuki validate                Validate registry manifests against the schema
  tanuki version                 Print the CLI version
  tanuki help                    Show this help

Catalog source: TANUKI_CATALOG_URL, ./catalog.json, ./dist/catalog.json,
or built directly from the registry/ manifests.
`)
}

func parseNameArg(cmd string, args []string, stderr io.Writer) (string, int) {
	if len(args) != 1 || strings.HasPrefix(args[0], "-") {
		fmt.Fprintf(stderr, "Usage: tanuki %s <service-name>\n", cmd)
		return "", 2
	}
	return args[0], 0
}

func parseSearchArgs(args []string, stderr io.Writer) (string, int) {
	team := ""
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--team":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				fmt.Fprintln(stderr, "--team requires a value")
				return "", 2
			}
			team = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--team="):
			team = strings.TrimPrefix(args[i], "--team=")
		default:
			fmt.Fprintf(stderr, "Unknown argument: %s\n", args[i])
			fmt.Fprintln(stderr, "Usage: tanuki search --team <team-name>")
			return "", 2
		}
	}
	if team == "" {
		fmt.Fprintln(stderr, "Usage: tanuki search --team <team-name>")
		return "", 2
	}
	return team, 0
}

func runList(stdout, stderr io.Writer) int {
	svcs, err := catalog.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	formatter.List(stdout, svcs)
	return 0
}

func runStatus(stdout, stderr io.Writer, name string) int {
	svcs, err := catalog.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	s := catalog.FindByName(svcs, name)
	if s == nil {
		fmt.Fprintf(stderr, "Service %q not found in catalog.\n", name)
		return 1
	}
	formatter.Status(stdout, *s)
	return 0
}

func runOwners(stdout, stderr io.Writer, name string) int {
	svcs, err := catalog.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	s := catalog.FindByName(svcs, name)
	if s == nil {
		fmt.Fprintf(stderr, "Service %q not found in catalog.\n", name)
		return 1
	}
	formatter.Owners(stdout, *s)
	return 0
}

func runSearch(stdout, stderr io.Writer, team string) int {
	svcs, err := catalog.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	filtered := catalog.FilterByTeam(svcs, team)
	formatter.List(stdout, filtered)
	if len(filtered) == 0 {
		return 1
	}
	return 0
}

func runValidate(stdout, stderr io.Writer) int {
	repoRoot, err := catalog.RepoRoot()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := schema.Validate(repoRoot, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
