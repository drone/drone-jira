// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"fmt"
	"regexp"
	"strings"
)

// helper function to extract the issue number from
// the commit details, including the commit message,
// branch and pull request title.
func extractIssue(args Args) string {
	return regexp.MustCompile(args.Project + "\\-\\d+").FindString(
		fmt.Sprintln(
			args.Commit.Message,
			args.PullRequest.Title,
			args.Commit.Source,
			args.Commit.Target,
		),
	)
}

// helper function determines the pipeline name.
func toPipeline(args Args) string {
	if v := args.Name; v != "" {
		return v
	}
	return args.Stage.Name
}

// helper function determines the pipeline state.
func toState(args Args) string {
	if v := args.State; v != "" {
		return toStateEnum(v)
	}
	return toStateEnum(args.Build.Status)
}

// helper function determines the target environment.
func toEnvironment(args Args) string {
	if v := args.Environment; v != "" {
		return toEnvironmentEnum(v)
	}
	if v := args.Deploy.Target; v != "" {
		return toEnvironmentEnum(v)
	}
	// default environment if none specified.
	return "production"
}

// helper function determines the version number.
func toVersion(args Args) string {
	if v := args.Semver.Version; v != "" {
		return v
	}
	if v := args.Tag.Name; v != "" {
		return v
	}
	return args.Commit.Rev
}

// helper function provides a deeplink to the build
// or a fallback link to the commit in version control.
func toLink(args Args) string {
	if v := args.Link; v != "" {
		return v
	}
	if v := args.Build.Link; v != "" {
		return v
	}
	return args.Commit.Link
}

// helper function normalizes the environment to match
// the expected bitbucket enum.
func toEnvironmentEnum(s string) string {
	switch strings.ToLower(s) {
	case "prod", "production":
		return "production"
	case "stage", "staging":
		return "staging"
	case "dev", "development":
		return "development"
	case "testing", "test":
		return "testing"
	default:
		return "unmapped"
	}
}

// helper function normalizes the state to match
// the expected bitbucket enum.
func toStateEnum(s string) string {
	switch strings.ToLower(s) {
	case "pending", "waiting":
		return "pending"
	case "running", "in_progress":
		return "in_progress"
	case "cancelled", "killed", "stopped", "terminated":
		return "cancelled"
	case "failed", "failure", "error", "errored":
		return "failed"
	case "rollback", "rolled_back":
		return "rolled_back"
	case "success", "successful":
		return "successful"
	default:
		return "unknown"
	}
}
