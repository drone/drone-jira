// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import "time"

type (
	BuildPayload struct {
		Builds []*Build `json:"builds"`
	}
	DeploymentPayload struct {
		Deployments []*Deployment `json:"deployments"`
	}

	// build provides the build details.
	Build struct {
		BuildNumber   int         `json:"buildNumber"`
		Description   string      `json:"description"`
		DisplayName   string      `json:"displayName"`
		IssueKeys     []string    `json:"issueKeys"`
		Label         string      `json:"label"`
		LastUpdated   time.Time   `json:"lastUpdated"`
		PipelineID    string      `json:"pipelineId"`
		References    []Reference `json:"references,omitempty"`
		SchemaVersion string      `json:"schemaVersion"`
		State         string      `json:"state"`
		TestInfo      struct {
			NumberFailed  int64 `json:"numberFailed"`
			NumberPassed  int64 `json:"numberPassed"`
			NumberSkipped int64 `json:"numberSkipped"`
			TotalNumber   int64 `json:"totalNumber"`
		} `json:"testInfo"`
		UpdateSequenceNumber int    `json:"updateSequenceNumber"`
		URL                  string `json:"url"`
	}
	Reference struct {
		Commit *CommitInfo `json:"commit,omitempty"` // Use a pointer to omit if nil
		Ref    *RefInfo    `json:"ref,omitempty"`    // Use a pointer to omit if nil
	}
	CommitInfo struct {
		ID            string `json:"id,omitempty"`
		RepositoryURI string `json:"repositoryUri,omitempty"`
	}

	RefInfo struct {
		Name string `json:"name,omitempty"`
		URI  string `json:"uri,omitempty"`
	}
	// Deployment provides the Deployment details.
	Deployment struct {
		Deploymentsequencenumber int `json:"deploymentSequenceNumber"`
		//IssueKeys                []string      `json:"issueKeys"`
		IssueKeys            []string      `json:"issueKeys,omitempty"`
		Updatesequencenumber int           `json:"updateSequenceNumber"`
		Associations         []Association `json:"associations"`
		Displayname          string        `json:"displayName"`
		URL                  string        `json:"url"`
		Description          string        `json:"description"`
		Lastupdated          time.Time     `json:"lastUpdated"`
		State                string        `json:"state"`
		Pipeline             JiraPipeline  `json:"pipeline"`
		Environment          Environment   `json:"environment"`
	}

	// Association provides the association details.
	Association struct {
		Associationtype string   `json:"associationType,omitempty"`
		Values          []string `json:"values,omitempty"`
	}

	// Environment provides the environment details.
	Environment struct {
		ID          string `json:"id"`
		Displayname string `json:"displayName"`
		Type        string `json:"type"`
	}

	// JiraPipeline provides the jira pipeline details.
	JiraPipeline struct {
		ID          string `json:"id"`
		Displayname string `json:"displayName"`
		URL         string `json:"url"`
	}

	// Tenant provides the jira instance tenant details.
	Tenant struct {
		ID string `json:"cloudId"`
	}

	// struct for adaptive card
	Card struct {
		Pipeline    string   `json:"pipeline"`
		Instance    string   `json:"instance"`
		Project     string   `json:"project"`
		State       string   `json:"state"`
		Version     string   `json:"version"`
		Environment string   `json:"environment"`
		URL         []string `json:"url"`
	}
)
