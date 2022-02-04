// Copyright 2021 Blend Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requireconditional

import (
	"context"
	"strings"
	"time"

	"github.com/blend/go-sdk/ex"
	"github.com/google/go-github/v42/github"
	githubactions "github.com/sethvargo/go-githubactions"
	yaml "gopkg.in/yaml.v3"

	"github.com/blend/require-conditional-status-checks/pkg/actions"
	githubshim "github.com/blend/require-conditional-status-checks/pkg/github"
)

// Config represents parsed configuration for this GitHub Action.
type Config struct {
	GitHubToken    string
	Timeout        time.Duration
	Interval       time.Duration
	ChecksYAML     string
	ChecksFilename string
	GitHubRootURL  string
	EventName      string
	EventAction    string
	GitHubOrg      string
	GitHubRepo     string
	BaseSHA        string
	HeadSHA        string
}

// NewFromInputs parses GitHub Actions inputs into a `Config`.
func NewFromInputs(action *githubactions.Action) (*Config, error) {
	timeout, err := actions.DurationInput(action, "timeout")
	if err != nil {
		return nil, err
	}

	interval, err := actions.DurationInput(action, "interval")
	if err != nil {
		return nil, err
	}

	c := Config{
		GitHubToken:    action.GetInput("github-token"),
		Timeout:        timeout,
		Interval:       interval,
		ChecksYAML:     action.GetInput("checks-yaml"),
		ChecksFilename: action.GetInput("checks-filename"),
	}
	err = c.setDefaults(action)
	if err != nil {
		return nil, err
	}

	return &c, c.Validate()
}

// setDefaults sets defaults which can be inferred from `GITHUB_*` environment
// variables. This is distinct from the explicit inputs to the action, which
// are provided as `INPUT_*` environment variables.
func (c *Config) setDefaults(action *githubactions.Action) error {
	c.GitHubRootURL = actions.RootURL(action)

	repository := actions.Repository(action)
	orgRepo := strings.SplitN(repository, "/", 2)
	if len(orgRepo) != 2 {
		return ex.New("Unexpected GitHub repository format", ex.OptMessagef("Repository: %q", repository))
	}
	c.GitHubOrg = orgRepo[0]
	c.GitHubRepo = orgRepo[1]

	// NOTE: Storing `c.EventName` duplicates a check that is also done in
	//       `actions.PullRequestEvent()`.
	c.EventName = actions.EventName(action)
	event, err := actions.PullRequestEvent(action)
	if err != nil {
		return err
	}
	if event != nil && event.Action != nil {
		c.EventAction = *event.Action
	}
	if event != nil && event.PullRequest != nil {
		if event.PullRequest.Base != nil {
			c.BaseSHA = event.PullRequest.Base.GetSHA()
		}
		if event.PullRequest.Head != nil {
			c.HeadSHA = event.PullRequest.Head.GetSHA()
		}
	}

	return nil
}

// Validate checks that a `Config` is valid.
// - The `EventName` must be `pull_request`
// - The `EventAction` must be `opened`, `synchronize` or `reopened`
// - The `BaseSHA` must be set to something
// - The `HeadSHA` must be set to something
// - The `GitHubOrg` must be set to something
// - The `GitHubRepo` must be set to something
// - The `GitHubRootURL` must be set to something
// - The `GitHubToken` must be set to something
// - Exactly one of `ChecksYAML` and `ChecksFilename` must be set
func (c Config) Validate() error {
	if c.EventName != "pull_request" {
		return ex.New("The Require Conditional Status Checks Action can only run on pull requests", ex.OptMessagef("Event Name: %q", c.EventName))
	}
	if !(c.EventAction == "opened" || c.EventAction == "synchronize" || c.EventAction == "reopened") {
		return ex.New("The Require Conditional Status Checks Action can only run on pull request types spawned by code changes", ex.OptMessagef("Event Action: %q", c.EventAction))
	}
	if c.BaseSHA == "" {
		return ex.New("Could not determine the base SHA for this pull request")
	}
	if c.HeadSHA == "" {
		return ex.New("Could not determine the head SHA for this pull request")
	}
	if c.GitHubOrg == "" {
		return ex.New("The Require Conditional Status Checks Action requires a GitHub owner or org")
	}
	if c.GitHubRepo == "" {
		return ex.New("The Require Conditional Status Checks Action requires a GitHub repository")
	}
	if c.GitHubRootURL == "" {
		return ex.New("The Require Conditional Status Checks Action requires a GitHub root URL")
	}
	if c.GitHubToken == "" {
		return ex.New("The Require Conditional Status Checks Action requires a GitHub API token")
	}
	if c.ChecksYAML != "" && c.ChecksFilename != "" {
		return ex.New("The Require Conditional Status Checks Action requires exactly one of checks YAML or checks filename; both are set")
	}
	if c.ChecksYAML == "" && c.ChecksFilename == "" {
		return ex.New("The Require Conditional Status Checks Action requires exactly one of checks YAML or checks filename; neither are set")
	}

	return nil
}

// GetChecks returns the checks for the current `Config`. Will be from either
// `ChecksYAML` or `ChecksFilename`. Using `ChecksFilename` will require a
// request to the GitHub API to read the file.
func (c Config) GetChecks(ctx context.Context, client *github.Client) ([]Check, error) {
	data, err := c.getChecksBytes(ctx, client)
	if err != nil {
		return nil, err
	}

	var checks []Check
	err = yaml.Unmarshal(data, &checks)
	if err != nil {
		return nil, ex.New("Failed to parse checks file as YAML", ex.OptInner(err))
	}

	return checks, nil
}

func (c Config) getChecksBytes(ctx context.Context, client *github.Client) ([]byte, error) {
	if c.ChecksYAML != "" {
		return []byte(c.ChecksYAML), nil
	}

	// NOTE: We assume `c.Validate()` has already passed, i.e. so
	//       `c.ChecksFilename` is set.
	return githubshim.GetFile(ctx, client, c.GitHubOrg, c.GitHubRepo, c.HeadSHA, c.ChecksFilename)
}
