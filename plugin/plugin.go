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
	// Environmnet Id (optional)
	EnvironmentId string `envconfig:"PLUGIN_ENVIRONMENT_ID"`
	// Environmnet Type (optional)
	EnvironmentType string `envconfig:"PLUGIN_ENVIRONMENT_TYPE"`

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
	// Issue Keys(optional)
	IssueKeys []string `envconfig:"PLUGIN_ISSUEKEYS"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	var (
		environ         = toEnvironment(args)
		environmentID   = toEnvironmentId(args)
		environmentType = toEnvironmentType(args)
		issues          []string
		state           = toState(args)
		version         = toVersion(args)
		deeplink        = toLink(args)
	)

	// ExtractInstanceName extracts the instance name from the provided URL if any
	instanceName := ExtractInstanceName(args.Instance)

	logger := logrus.
		WithField("client_id", args.ClientID).
		WithField("cloud_id", args.CloudID).
		WithField("project_id", args.Project).
		WithField("instance", instanceName).
		WithField("pipeline", args.Name).
		WithField("environment", environ).
		WithField("state", state).
		WithField("environment Type", environmentType).
		WithField("environment ID", environmentID)

	// check if PLUGIN_ISSUEKEYS is provided
	if len(args.IssueKeys) > 0 {
		logger.Debugln("Provided issue keys are :", args.IssueKeys)
		issues = args.IssueKeys
	} else {
		// fallback to extracting from commit if no issue keys are passed
		issues = extractIssues(args)
		if len(issues) == 0 {
			logger.Debugln("cannot find issue number")
			return errors.New("failed to extract issue number")
		}
	}
	logger = logger.WithField("issues", strings.Join(issues, ","))
	logger.Debugln("successfully extracted all issues")

	commitMessage := args.Commit.Message
	if len(commitMessage) > 255 {
		logger.Warnln("Commit message exceeds 255 characters; truncating to fit.")
		commitMessage = commitMessage[:252] + "..."
	}

	logger.Debugln("successfully extraced issue number")
	deploymentPayload := DeploymentPayload{
		Deployments: []*Deployment{
			{
				Deploymentsequencenumber: args.Build.Number,
				Updatesequencenumber:     args.Build.Number,
				Associations: []Association{
					{
						Associationtype: "issueIdOrKeys",
						Values:          issues,
					},
				},
				Displayname: strconv.Itoa(args.Build.Number),
				URL:         deeplink,
				Description: commitMessage,
				Lastupdated: time.Now(),
				State:       state,
				Pipeline: JiraPipeline{
					ID:          args.Name,
					Displayname: args.Name,
					URL:         deeplink,
				},
				Environment: Environment{
					ID:          environmentID,
					Displayname: environ,
					Type:        environmentType,
				},
			},
		},
	}
	// Initialize an empty reference
	references := []Reference{}
	// Check if any input is available to update the reference
	if args.Commit.Branch != "" || args.Commit.Link != "" || args.Commit.Rev != "" {
		var reference Reference
		// Update CommitInfo if Rev or Link is provided
		if args.Commit.Rev != "" || args.Commit.Link != "" {
			reference.Commit = &CommitInfo{
				ID:            args.Commit.Rev,
				RepositoryURI: args.Commit.Link,
			}
		}
		// Update RefInfo if both Branch and Link are provided
		if args.Commit.Branch != "" && args.Commit.Link != "" {
			reference.Ref = &RefInfo{
				Name: args.Commit.Branch,
				URI:  fmt.Sprintf("%s/refs/%s", args.Commit.Link, args.Commit.Branch),
			}
		}

		// Append the reference if at least one field is populated
		if reference.Commit != nil || reference.Ref != nil {
			references = append(references, reference)
		}
	}
	// Build the Build struct and include references only if non-empty
	buildPayload := BuildPayload{
		Builds: []*Build{
			{
				BuildNumber:          args.Build.Number,
				Description:          commitMessage,
				DisplayName:          args.Name,
				URL:                  deeplink,
				LastUpdated:          time.Now(),
				PipelineID:           args.Name,
				IssueKeys:            issues,
				State:                state,
				UpdateSequenceNumber: args.Build.Number,
				References:           references,
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
		cloudID, err := getCloudID(instanceName, args.CloudID)
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
		logger.Infoln("creating deployment")
		deploymentErr := createDeployment(deploymentPayload, cloudID, args.Level, oauthToken)
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
		if args.EnvironmentName != "" {
			logger.Infoln("creating deployment")
			deploymentErr := createConnectDeployment(deploymentPayload, instanceName, args.Level, jwtToken)
			if deploymentErr != nil {
				logger.WithError(deploymentErr).
					Errorln("cannot create deployment")
				return deploymentErr
			}
		} else {
			logger.Infoln("creating build")
			buildErr := createConnectBuild(buildPayload, instanceName, args.Level, jwtToken)
			if buildErr != nil {
				logger.WithError(buildErr).
					Errorln("cannot create build")
				return buildErr
			}
		}
	}
	// only create card if the state is successful

	var ticketLinks []string

	if len(issues) > 0 {
		for _, issue_key := range issues {
			ticketLink := fmt.Sprintf("https://%s.atlassian.net/browse/%s", args.Instance, issue_key)
			ticketLinks = append(ticketLinks, ticketLink)
		}
	}
	cardData := Card{
		Pipeline:    args.Name,
		Instance:    instanceName,
		Project:     args.Project,
		State:       state,
		Version:     version,
		Environment: environ,
		URL:         ticketLinks,
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
		return "", fmt.Errorf("errorCode %d", res.StatusCode)
	}
	output := map[string]interface{}{}
	err = json.Unmarshal(out, &output)
	if err != nil {
		return "", err
	}
	//	fmt.Println(output["access_token"].(string))
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
func createDeployment(payload DeploymentPayload, cloudID, debug, oauthToken string) error {
	endpoint := fmt.Sprintf("https://api.atlassian.com/jira/deployments/0.1/cloud/%s/bulk", cloudID)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return err
	}
	req.Header.Set("From", "noreply@localhost")
	req.Header.Set("Authorization", "Bearer "+oauthToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
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
		return fmt.Errorf("errorCode %d", res.StatusCode)
	}
	return nil
}

// makes an API call to create a deployment.
func createConnectDeployment(payload DeploymentPayload, cloudID, debug, jwtToken string) error {
	endpoint := fmt.Sprintf("https://%s.atlassian.net/rest/deployments/0.1/bulk", cloudID)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return err
	}
	req.Header.Set("From", "noreply@localhost")
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
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
		return fmt.Errorf("errorCode %d", res.StatusCode)
	}
	return nil
}

// makes an API call to create a build.
func createConnectBuild(payload BuildPayload, cloudID, debug, jwtToken string) error {
	endpoint := fmt.Sprintf("https://%s.atlassian.net/rest/builds/0.1/bulk", cloudID)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return err
	}
	req.Header.Set("From", "noreply@localhost")
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
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
		return fmt.Errorf("errorCode %d", res.StatusCode)
	}
	return nil
}

func getCloudID(instance, cloudID string) (string, error) {
	if instance != "" {

		tenant, err := lookupTenant(instance)
		if err != nil {
			return "", fmt.Errorf("cannotGetCloudIdFromInstance, %s", err)
		}
		return tenant.ID, nil
	}
	if cloudID == "" {
		return "", fmt.Errorf("cloudIdIsEmptySpecifyTheCloudIdOrInstanceName")
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
		return nil, fmt.Errorf("errorCode %d", res.StatusCode)
	}
	out := new(Tenant)
	err = json.NewDecoder(res.Body).Decode(out)
	return out, err
}
