# drone-jira

A plugin to attach build and deployment details to a Jira issue. For information on how to use the plugin with drone there is a video [here](https://youtu.be/YIKbLeY1-gI)

This plugin was modified to extract multiple issues by changing the extractIssues function and slight changes to the file plugin.go, also was added a new argument AdditionalMessage to serve as an input variable.
Changed the util_test.go file to suit the modifications.

## Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t plugins/jira -f docker/Dockerfile .
```

## Testing

Execute the plugin from your current working directory:

```text
docker run --rm \
  -e DRONE_COMMIT_SHA=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_COMMIT_AUTHOR=octocat \
  -e DRONE_COMMIT_AUTHOR_EMAIL=octocat@github.com \
  -e DRONE_COMMIT_MESSAGE="DRONE-42 updated the readme" \
  -e DRONE_BUILD_NUMBER=43 \
  -e DRONE_BUILD_STATUS=success \
  -e DRONE_BUILD_LINK=https://cloud.drone.io \
  -e PLUGIN_CLOUD_ID=${JIRA_CLOUD_ID} \
  -e PLUGIN_CLIENT_ID=${JIRA_CLIENT_ID} \
  -e PLUGIN_CLIENT_SECRET=${JIRA_CLIENT_SECRET} \
  -e PLUGIN_PROJECT=${JIRA_PROJECT} \
  -e PLUGIN_PIPELINE=drone \
  -e PLUGIN_ENVIRONMENT=production \
  -e PLUGIN_STATE=successful \
  -w /drone/src \
  -v $(pwd):/drone/src \
  plugins/jira
```

## Plugin Settings
- `LOG_LEVEL` debug/info Level defines the plugin log level. Set this to debug to see the response from jira
- `CLOUD_ID` Atlassian Cloud ID (required)
- `CLIENT_ID` Atlassian Oauth2 Client ID (required)
- `CLIENT_SECRET` Atlassian Oauth2 Client Secret (required)
- `INSTANCE` Site Name (optional)
- `PROJECT` Project Name (required)
- `PIPELINE` Pipeline Name (optional)
- `ENVIRONMENT_NAME` Deployment environment (optional)
- `LINK` Link to deployment (optional)
- `STATE` State of the deployment (optional)
	
