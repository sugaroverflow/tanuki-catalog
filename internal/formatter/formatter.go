package formatter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gitlab.com/gitlab-da/projects/tanuki-services/tanuki-catalog/internal/catalog"
)

// List writes a table of services (name, version, team, owner) to w.
func List(w io.Writer, services []catalog.Service) {
	if len(services) == 0 {
		fmt.Fprintln(w, "No services found.")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tVERSION\tTEAM\tOWNER")
	for _, s := range services {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.Name, s.Version, s.Team, s.Owner)
	}
	_ = tw.Flush()
}

// Status writes detailed status for one service to w.
func Status(w io.Writer, s catalog.Service) {
	fmt.Fprintf(w, "Name:        %s\n", s.Name)
	fmt.Fprintf(w, "Version:     %s\n", s.Version)
	fmt.Fprintf(w, "Owner:       %s\n", s.Owner)
	fmt.Fprintf(w, "Team:        %s\n", s.Team)
	fmt.Fprintf(w, "Health URL:  %s\n", s.HealthURL)
	fmt.Fprintf(w, "Repo:        %s\n", s.RepoURL)
	fmt.Fprintf(w, "Last deploy: %s\n", s.LastDeploy)
	if s.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", s.Description)
	}
}

// Owners writes owner and on-call info for one service to w.
func Owners(w io.Writer, s catalog.Service) {
	fmt.Fprintf(w, "Service:  %s\n", s.Name)
	fmt.Fprintf(w, "Owner:    %s\n", s.Owner)
	if len(s.Owners) > 0 {
		fmt.Fprintf(w, "Others:   %s\n", strings.Join(s.Owners, ", "))
	}
	if s.OnCall != "" {
		fmt.Fprintf(w, "On-call:  %s\n", s.OnCall)
	} else {
		fmt.Fprintln(w, "On-call:  (none listed)")
	}
}
