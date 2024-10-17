// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	filepath "github.com/bmatcuk/doublestar"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/drone/go-scm/scm/transport"
	"github.com/sirupsen/logrus"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Include patterns to check
	Include []string `envconfig:"PLUGIN_INCLUDE"`
	// Exclude patterns to check
	Exclude []string `envconfig:"PLUGIN_EXCLUDE"`

	// GitHub API Token
	GithubToken string `envconfig:"PLUGIN_GITHUB_TOKEN"`
	//
	GithubServer string `envconfig:"PLUGIN_GITHUB_SERVER"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	// matchSeen will be true if a file changed in the commit range
	// matches the include/exclude pattern(s)
	matchSeen := false

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	err := validate(&args)
	if err != nil {
		return err
	}

	files, err := getGithubFilesChanged(ctx, &args)
	if err != nil {
		return err
	}

	if len(files) > 0 {
		for _, file := range files {
			got, want := match(args.Exclude, args.Include, file), true
			if got == want {
				logrus.Infoln("match seen for file", file)
				matchSeen = true
			}
		}
	}

	writeOutput(&args, "MATCH_SEEN", strconv.FormatBool(matchSeen))

	return nil
}

func validate(args *Args) error {
	// variables set by plugin parameters
	if args.GithubToken == "" {
		return fmt.Errorf("missing github token")
	}
	if len(args.Include) == 0 && len(args.Exclude) == 0 {
		return fmt.Errorf("include or exclude must be set")
	}

	// variables automatically set by Drone/Harness
	if args.Output.Name == "" {
		return fmt.Errorf("DRONE_OUTPUT is unset")
	}
	if args.Pipeline.Commit.Before == "" {
		return fmt.Errorf("DRONE_COMMIT_BEFORE is unset")
	}
	if args.Pipeline.Commit.After == "" {
		return fmt.Errorf("DRONE_COMMIT_AFTER is unset")
	}
	if args.Pipeline.Repo.Slug == "" {
		return fmt.Errorf("DRONE_REPO is unset")
	}

	return nil
}

// getGithubFilesChanged gets a list of files changed between before and after
// commits that triggered the pipeline
func getGithubFilesChanged(ctx context.Context, args *Args) ([]string, error) {
	var client *scm.Client
	var err error

	if args.GithubServer == "" {
		client = github.NewDefault()
	} else {
		client, err = github.New(args.GithubServer + "/api/v3")
		if err != nil {
			return nil, err
		}
	}

	client.Client = &http.Client{
		Transport: &transport.BearerToken{
			Token: args.GithubToken,
		},
	}

	var changes []*scm.Change
	var result *scm.Response

	if args.Pipeline.Commit.Before == "" || args.Pipeline.Commit.Before == scm.EmptyCommit {
		changes, result, err = client.Git.ListChanges(ctx, args.Pipeline.Repo.Slug, args.Pipeline.Commit.After, scm.ListOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		changes, result, err = client.Git.CompareChanges(ctx, args.Pipeline.Repo.Slug, args.Pipeline.Commit.Before, args.Pipeline.Commit.After, scm.ListOptions{})
		if err != nil {
			return nil, err
		}
	}

	logrus.Debugln("token API calls per hour remaining:", result.Rate.Remaining)

	var files []string
	for _, c := range changes {
		files = append(files, c.Path)
	}

	return files, nil
}

// match returns true if the string matches the include
// patterns and does not match any of the exclude patterns.
func match(exclude []string, include []string, file string) bool {
	if excludes(exclude, file) {
		return false
	}
	if includes(include, file) {
		return true
	}
	if len(include) == 0 {
		return true
	}
	return false
}

// includes returns true if the string matches the include
// patterns.
func includes(patterns []string, v string) bool {
	for _, pattern := range patterns {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// excludes returns true if the string matches the exclude
// patterns.
func excludes(patterns []string, v string) bool {
	for _, pattern := range patterns {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// writeOutput writes a variable and its value to the output file.
func writeOutput(args *Args, key, value string) error {
	outputFile, err := os.OpenFile(args.Output.Name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer outputFile.Close()

	_, err = fmt.Fprintf(outputFile, "%s=%s\n", key, value)
	if err != nil {
		return fmt.Errorf("failed to write to env: %w", err)
	}

	return nil
}
