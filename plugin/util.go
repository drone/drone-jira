// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
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
			args.Commit.Branch,
		),
	)
}

// helper function determines the pipeline state.
func toState(args Args) string {
	if v := args.State; v != "" {
		return toStateEnum(v)
	}
	return toStateEnum(args.Build.Status)
}

// helper function determines the target environment Name.
func toEnvironment(args Args) string {
	if v := args.EnvironmentName; v != "" {
		return toEnvironmentEnum(v)
	}
	if v := args.Deploy.Target; v != "" {
		return toEnvironmentEnum(v)
	}
	// default environment if none specified.
	return "production"
}

// helper function determines the target environment Id.
func toEnvironmentId(args Args) string {
	if v := args.EnvironmentId; v != "" {
		return v
	}
	// Return a default value, such as an empty string
	return toEnvironment(args)
}

// helper function determines the target environment Type.
func toEnvironmentType(args Args) string {
	if v := args.EnvironmentType; v != "" {
		return v
	}
	// Return a default value, such as an empty string
	return ""
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

// helper function ExtractInstanceName extracts the instance name from the provided URL
// or returns the instance name directly
func ExtractInstanceName(instance string) string {
	// Check if the instance is a full URL
	if strings.Contains(instance, "://") {
		parsedURL, err := url.Parse(instance)
		if err == nil {
			// Return the host part without the top-level domain
			hostParts := strings.Split(parsedURL.Hostname(), ".")
			if len(hostParts) > 0 {
				return hostParts[0] // Return the first part as the instance name
			}
		} else {
			// Log the error if URL parsing fails
			logrus.WithField("instance", instance).WithField("err", err).Error("Error parsing URL")
		}
	} else {
		// If it's not a URL, split by dots to get the instance name
		hostParts := strings.Split(instance, ".")
		if len(hostParts) > 0 {
			return hostParts[0] // Return the first part as the instance name
		}
	}
	// Default return if no valid instance name is found
	return instance
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
