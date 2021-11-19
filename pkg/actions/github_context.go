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

package actions

import (
	"encoding/json"
	"os"

	"github.com/blend/go-sdk/ex"
	"github.com/google/go-github/v40/github"
	githubactions "github.com/sethvargo/go-githubactions"
)

// PullRequestEvent parses the `${{ github.event }}` from the filesystem.
// Since actions code doesn't have direct access to the `${{ github }}`
// context, the `GITHUB_EVENT_PATH` environment variable is used to read the
// event. If the `${{ github.event_name }}` is **not** pull request, this
// will fail.
func PullRequestEvent(action *githubactions.Action) (*github.PullRequestEvent, error) {
	eventName := EventName(action)
	if eventName != "pull_request" {
		return nil, ex.New("GitHub Action not triggered for a pull request", ex.OptMessagef("Event Name: %s", eventName))
	}

	githubEventPath := actionGetEnv(action, "GITHUB_EVENT_PATH")
	githubEventBytes, err := os.ReadFile(githubEventPath)
	if os.IsNotExist(err) {
		return nil, ex.New("GitHub Event file does not exist", ex.OptMessagef("Path: %s", githubEventPath))
	}
	if err != nil {
		return nil, ex.New("Could not read GitHub Event file", ex.OptMessagef("Path: %s", githubEventPath), ex.OptInner(err))
	}

	var event github.PullRequestEvent
	err = json.Unmarshal(githubEventBytes, &event)
	if err != nil {
		return nil, ex.New("Failed to parse GitHub Event file as JSON", ex.OptMessagef("Path: %s", githubEventPath), ex.OptInner(err))
	}

	return &event, nil
}

// EventName returns the `${{ github.event_name }}` context value.
func EventName(action *githubactions.Action) string {
	return actionGetEnv(action, "GITHUB_EVENT_NAME")
}

// Repository returns the `${{ github.repository }}` context value.
func Repository(action *githubactions.Action) string {
	return actionGetEnv(action, "GITHUB_REPOSITORY")
}

// RootURL returns the `${{ github.api_url }}` context value. For public
// GitHub this is expected to be `https://api.github.com` and for GitHub
// Enterprise it is expected to be `https://[hostname]/api/v3`.
func RootURL(action *githubactions.Action) string {
	return actionGetEnv(action, "GITHUB_API_URL")
}
