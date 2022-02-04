// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import "time"

type (
	// Payload provides the Deployment payload.
	Payload struct {
		Deployments []*Deployment `json:"deployments"`
	}

	// Deployment provides the Deployment details.
	Deployment struct {
		Deploymentsequencenumber int           `json:"deploymentSequenceNumber"`
		Updatesequencenumber     int           `json:"updateSequenceNumber"`
		Associations             []Association `json:"associations"`
		Displayname              string        `json:"displayName"`
		URL                      string        `json:"url"`
		Description              string        `json:"description"`
		Lastupdated              time.Time     `json:"lastUpdated"`
		State                    string        `json:"state"`
		Pipeline                 JiraPipeline  `json:"pipeline"`
		Environment              Environment   `json:"environment"`
	}

	// Association provides the association details.
	Association struct {
		Associationtype string   `json:"associationType"`
		Values          []string `json:"values"`
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
		Pipeline    string `json:"pipeline"`
		Instance    string `json:"instance"`
		Project     string `json:"project"`
		State       string `json:"state"`
		Version     string `json:"version"`
		Environment string `json:"environment"`
		URL         string `json:"url"`
	}
)
