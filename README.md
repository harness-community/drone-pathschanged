Drone plugin to detect files changed in a commit range.

This goal of this plugin is to provide an alternative to the [pathschanged conversion extension](https://github.com/meltwater/drone-convert-pathschanged) as a pipeline plugin.

# Usage

NOTE: This plugin writes to [DRONE_OUTPUT](https://developer.harness.io/docs/continuous-integration/troubleshoot-ci/ci-env-var/#drone_output) which is a feature of [Harness CI](https://www.harness.io/products/continuous-integration). The `DRONE_OUTPUT` file will contain a `MATCH_SEEN` variable which will be `true` if a match was seen, and `false` otherwise.

To use this plugin in a Drone pipeline, you must manage the `DRONE_OUTPUT` file and variable yourself.

This extension uses [doublestar](https://github.com/bmatcuk/doublestar) for matching paths changed in your commit range, refer to their documentation for all supported patterns.

The following settings changes this plugin's behavior.

* `github_token`: github personal access token (required)
* `github_server`: github server URL (optional)
* `include`: one or more file patterns (must specify this and/or `exclude`)
* `exclude`: one or more file patterns (must specify this and/or `include`)

Below is an example `.drone.yml` that uses this plugin.

```yaml
kind: pipeline
name: default

steps:
- name: run pathschanged plugin
  image: jimsheldon/drone-pathschanged
  pull: if-not-exists
  environment:
    DRONE_OUTPUT: output_variables.txt
  settings:
    github_token:
       from_secret: github_pat
    include: README.md
```

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t jimsheldon/drone-pathschanged -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
docker run --rm \
  -e PLUGIN_GITHUB_TOKEN=$GITHUB_PAT \
  -e PLUGIN_INCLUDE=README.md \
  -e DRONE_REPO=octocat/Hello-World \
  -e DRONE_OUTPUT=output_variables.txt \
  -e DRONE_COMMIT_BEFORE=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_AFTER=44eb4b38b6ccd42b667ddc6a4d5226c842e8fab7 \
  -w /drone/src \
  -v $(pwd):/drone/src \
  jimsheldon/drone-pathschanged
```
