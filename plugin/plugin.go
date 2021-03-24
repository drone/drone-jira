// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// TODO code to generate cloud id (if not exists)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Atlassian Cloud ID (required)
	CloudID string `envconfig:"PLUGIN_CLOUD_ID"`

	// Atlassian Oauth2 Client ID (required)
	ClientID string `envconfig:"PLUGIN_CLIENT_ID"`

	// Atlassian Oauth2 Client Secret (required)
	ClientSecret string `envconfig:"PLUGIN_CLIENT_SECRET"`

	// Project Name (required)
	Project string `envconfig:"PLUGIN_PROJECT"`

	// Pipeline Name (optional)
	Name string `envconfig:"PLUGIN_PIPELINE"`

	// Deployment environment (optional)
	Environment string `envconfig:"PLUGIN_ENVIRONMENT"`

	// Link to deployment (optional)
	Link string `envconfig:"PLUGIN_LINK"`

	// State of the deployment (optional)
	State string `envconfig:"PLUGIN_STATE"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	var (
		environ  = toEnvironment(args)
		issue    = extractIssue(args)
		state    = toState(args)
		version  = toVersion(args)
		deeplink = toLink(args)
	)

	logger := logrus.
		WithField("client_id", args.ClientID).
		WithField("cloud_id", args.CloudID).
		WithField("project_id", args.Project).
		WithField("pipeline", args.Name).
		WithField("environment", environ).
		WithField("state", state).
		WithField("version", version)

	if issue == "" {
		logger.Debugln("cannot find issue number")
		return errors.New("failed to extract issue number")
	}

	logger = logger.WithField("issue", issue)
	logger.Debugln("successfully extraced issue number")

	payload := Payload{
		Deployments: []*Deployment{
			{
				Deploymentsequencenumber: args.Build.Number,
				Updatesequencenumber:     args.Build.Number,
				Associations: []Association{
					{
						Associationtype: "issueIdOrKeys",
						Values:          []string{issue},
					},
				},
				Displayname: version,
				URL:         deeplink,
				Description: args.Commit.Message,
				Lastupdated: time.Now(),
				State:       state,
				Pipeline: JiraPipeline{
					ID:          args.Commit.Author.Email,
					Displayname: args.Commit.Author.Username,
					URL:         deeplink,
				},
				Environment: Environment{
					ID:          environ,
					Displayname: environ,
					Type:        environ,
				},
			},
		},
	}

	logrus.Debugln("creating token")
	token, err := createToken(args)
	if err != nil {
		logrus.Debugln("cannot create token")
		return err
	}

	logrus.Debugln("creating deployment")
	return createDeployment(args, payload, token)
}

// makes an API call to create a token.
func createToken(args Args) (string, error) {
	payload := map[string]string{
		"audience":      "api.atlassian.com",
		"grant_type":    "client_credentials",
		"client_id":     args.ClientID,
		"client_secret": args.ClientSecret,
	}
	endpoint := "https://api.atlassian.com/oauth/token"
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(payload)
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode > 299 {
		return "", fmt.Errorf("Error code %d", res.StatusCode)
	}
	output := map[string]interface{}{}
	err = json.Unmarshal(out, &output)
	if err != nil {
		return "", err
	}
	return output["access_token"].(string), nil
}

// makes an API call to create a deployment.
func createDeployment(args Args, payload Payload, token string) error {
	endpoint := fmt.Sprintf("https://api.atlassian.com/jira/deployments/0.1/cloud/%s/bulk", args.CloudID)
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(payload)
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return err
	}
	req.Header.Set("From", "noreply@localhost")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		return fmt.Errorf("Error code %d", res.StatusCode)
	}
	return nil
}
