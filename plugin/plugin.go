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
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"

	"github.com/sirupsen/logrus"
)

const (
	// DefaultConnectHostname is the default connect hostname
	DefaultConnectHostname = "https://jira-ci.harness.io"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Atlassian Cloud ID (required)
	CloudID string `envconfig:"PLUGIN_CLOUD_ID"`

	// Instance Name (optional)
	Instance string `envconfig:"PLUGIN_INSTANCE"`

	// Project Name (required)
	Project string `envconfig:"PLUGIN_PROJECT"`

	// Pipeline Name (required)
	Name string `envconfig:"PLUGIN_PIPELINE"`

	// Deployment environment (optional)
	EnvironmentName string `envconfig:"PLUGIN_ENVIRONMENT_NAME"`

	// Link to deployment (optional)
	Link string `envconfig:"PLUGIN_LINK"`

	// State of the deployment (optional)
	State string `envconfig:"PLUGIN_STATE"`

	// Path to the adaptive card
	CardFilePath string `envconfig:"DRONE_CARD_PATH"`

	// AUTHENTICATION
	// Atlassian Oauth Client ID (required)
	ClientID string `envconfig:"PLUGIN_CLIENT_ID"`

	// Atlassian Oauth2 Client Secret (required)
	ClientSecret string `envconfig:"PLUGIN_CLIENT_SECRET"`

	// Connect KEY (required) - if client id and secret are not provided
	ConnnectKey string `envconfig:"PLUGIN_CONNECT_KEY"`

	// connect hostname (required)
	ConnectHostname string `envconfig:"PLUGIN_CONNECT_HOSTNAME"`
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
		WithField("instance", args.Instance).
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

	deploymentPayload := DeploymentPayload{
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
				Displayname: strconv.Itoa(args.Build.Number),
				URL:         deeplink,
				Description: args.Commit.Message,
				Lastupdated: time.Now(),
				State:       state,
				Pipeline: JiraPipeline{
					ID:          args.Name,
					Displayname: args.Name,
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
	buildPayload := BuildPayload{
		Builds: []*Build{
			{
				BuildNumber:          args.Build.Number,
				Description:          args.Commit.Message,
				DisplayName:          args.Name,
				URL:                  deeplink,
				LastUpdated:          time.Now(),
				PipelineID:           args.Name,
				IssueKeys:            []string{issue},
				State:                state,
				UpdateSequenceNumber: args.Build.Number,
			},
		},
	}

	// validation of arguments
	if (args.ClientID == "" && args.ClientSecret == "") && (args.ConnnectKey == "") {
		logger.Debugln("client id and secret are empty. specify the client id and secret or specify connect key")
		return errors.New("No client id & secret or connect token & hostname provided")
	}
	// create tokens and deployments
	if args.ClientID != "" && args.ClientSecret != "" {
		// get cloud id
		cloudID, err := getCloudID(args.Instance, args.CloudID)
		if err != nil {
			logger.Debugln("cannot get cloud id")
			return err
		}
		logger.Debugln("creating oauth token for deployment")
		oauthToken, err := getOauthToken(args)
		if err != nil {
			logger.Debugln("cannot create token, from client id and secret")
			return err
		}
		authTransport := &AuthTransport{Token: oauthToken}
		closed, err := isJiraIssueClosed(ctx, authTransport, cloudID, issue)
		if err != nil {
			logger.WithError(err).
				Errorln("cannot check if issue is closed")
			return err
		}

		if closed {
			return fmt.Errorf("issue: %s is closed", issue)
		}

		logger.Infoln("creating deployment")
		deploymentErr := createDeployment(deploymentPayload, authTransport, cloudID, args.Level)
		if deploymentErr != nil {
			logger.WithError(deploymentErr).
				Errorln("cannot create deployment")
			return deploymentErr
		}
	} else {
		// set default connect hostname
		if args.ConnectHostname == "" {
			args.ConnectHostname = DefaultConnectHostname
		}
		logger.Debugln("creating jwt token from connect key")
		jwtToken, err := getConnectToken(args.ConnnectKey, args.ConnectHostname)
		if err != nil {
			logger.Debugln("cannot get jwt token, from connect key")
			return err
		}

		authTransport := &AuthTransport{Token: jwtToken}
		closed, err := isJiraIssueClosed(ctx, authTransport, args.Instance, issue)
		if err != nil {
			logger.WithError(err).
				Errorln("cannot check if issue is closed")
			return err
		}

		if closed {
			return fmt.Errorf("issue: %s is closed", issue)
		}

		if args.EnvironmentName != "" {
			logger.Infoln("creating deployment")
			deploymentErr := createConnectDeployment(deploymentPayload, authTransport, args.Instance, args.Level)
			if deploymentErr != nil {
				logger.WithError(deploymentErr).
					Errorln("cannot create deployment")
				return deploymentErr
			}
		} else {
			logger.Infoln("creating build")
			buildErr := createConnectBuild(buildPayload, authTransport, args.Instance, args.Level)
			if buildErr != nil {
				logger.WithError(buildErr).
					Errorln("cannot create build")
				return buildErr
			}
		}
	}
	// only create card if the state is successful
	ticketLink := fmt.Sprintf("https://%s.atlassian.net/browse/%s", args.Instance, issue)
	cardData := Card{
		Pipeline:    args.Name,
		Instance:    args.Instance,
		Project:     args.Project,
		State:       state,
		Version:     version,
		Environment: environ,
		URL:         ticketLink,
	}
	if err := args.writeCard(cardData); err != nil {
		fmt.Printf("Could not create adaptive card. %s\n", err)
		return err
	}
	return nil
}

// makes an API call to create a token.
func getOauthToken(args Args) (string, error) {
	payload := map[string]string{
		"audience":      "api.atlassian.com",
		"grant_type":    "client_credentials",
		"client_id":     args.ClientID,
		"client_secret": args.ClientSecret,
	}
	endpoint := "https://api.atlassian.com/oauth/token"
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return "", err
	}
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

	out, err := io.ReadAll(res.Body)
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

func getConnectToken(connectToken, connectURL string) (token string, err error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/token", connectURL), nil)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", connectToken))

	res, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		return "", httpErr
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	// strip characters from the response
	jwtString := string(body)
	return jwtString, nil
}

// makes an API call to create a deployment.
func createDeployment(payload DeploymentPayload, authTransport *AuthTransport, cloudID, debug string) error {
	endpoint := fmt.Sprintf("https://api.atlassian.com/jira/deployments/0.1/cloud/%s/bulk", cloudID)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return err
	}
	res, err := authTransport.Client().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch debug {
	case "debug", "trace", "DEBUG", "TRACE":
		out, _ := httputil.DumpResponse(res, true)
		outString := string(out)
		logrus.WithField("status", res.Status).WithField("response", outString).Info("request complete")
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("Error code %d", res.StatusCode)
	}
	return nil
}

// makes an API call to create a deployment.
func createConnectDeployment(payload DeploymentPayload, authTransport *AuthTransport, cloudID, debug string) error {
	endpoint := fmt.Sprintf("https://%s.atlassian.net/rest/deployments/0.1/bulk", cloudID)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return err
	}
	res, err := authTransport.Client().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch debug {
	case "debug", "trace", "DEBUG", "TRACE":
		out, _ := httputil.DumpResponse(res, true)
		outString := string(out)
		logrus.WithField("status", res.Status).WithField("response", outString).Info("request complete")
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("Error code %d", res.StatusCode)
	}
	return nil
}

// makes an API call to create a build.
func createConnectBuild(payload BuildPayload, authTransport *AuthTransport, cloudID, debug string) error {
	endpoint := fmt.Sprintf("https://%s.atlassian.net/rest/builds/0.1/bulk", cloudID)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return err
	}
	res, err := authTransport.Client().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch debug {
	case "debug", "trace", "DEBUG", "TRACE":
		out, _ := httputil.DumpResponse(res, true)
		outString := string(out)
		logrus.WithField("status", res.Status).WithField("response", outString).Info("request complete")
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("Error code %d", res.StatusCode)
	}
	return nil
}

func getCloudID(instance, cloudID string) (string, error) {
	if instance != "" {

		tenant, err := lookupTenant(instance)
		if err != nil {
			return "", fmt.Errorf("Cannot get cloudid from instance, %s", err)
		}
		return tenant.ID, nil
	}
	if cloudID == "" {
		return "", fmt.Errorf("cloud id is empty. specify the cloud id or instance name")
	}
	return cloudID, nil
}

// makes an API call to lookup the cloud ID
func lookupTenant(tenant string) (*Tenant, error) {
	uri := fmt.Sprintf("https://%s.atlassian.net/_edge/tenant_info", tenant)
	res, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		return nil, fmt.Errorf("Error code %d", res.StatusCode)
	}
	out := new(Tenant)
	err = json.NewDecoder(res.Body).Decode(out)
	return out, err
}

func isJiraIssueClosed(ctx context.Context, authTransport *AuthTransport, cloudID, issueID string) (bool, error) {

	logger := logrus.WithField("issueID", issueID)
	endpoint := fmt.Sprintf("https://%s.atlassian.net/", cloudID)

	jiraClient, err := jira.NewClient(endpoint, authTransport.Client())
	if err != nil {
		logger.WithError(err).
			Errorln("cannot connect to jira.")
		return true, err
	}

	issue, resp, err := jiraClient.Issue.Get(ctx, issueID, nil)
	if err != nil {
		fmt.Println(resp.Status)
		logger.WithError(err).
			Errorln("cannot get issue")
		return true, err
	}

	if issue.Fields.Status == nil {
		logger.Debug("no status found on issue")
		return true, nil
	}

	if strings.ToUpper(issue.Fields.Status.Name) == "CLOSED" {
		logger.Debug("issue is closed")
		return true, nil
	}

	logger.Debug("issue is not closed it is in state", issue.Fields.Status.Name)
	return false, nil
}
