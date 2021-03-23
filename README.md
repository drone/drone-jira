A plugin to attach build and deployment details to a Jira issue.

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t plugins/jira -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
docker run --rm \
  -e DRONE_COMMIT_SHA=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_COMMIT_AUTHOR=bradrydzewski \
  -e DRONE_COMMIT_AUTHOR_EMAIL=brad.rydzewski@gmail.com \
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