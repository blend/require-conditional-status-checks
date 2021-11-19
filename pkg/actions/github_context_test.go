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

package actions_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/blend/go-sdk/assert"
	githubactions "github.com/sethvargo/go-githubactions"

	"github.com/blend/action-composite/pkg/actions"
)

func TestPullRequestEvent(t *testing.T) {
	t.Parallel()
	it := assert.New(t)

	// 1. Not a `pull_request`
	action := githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_EVENT_PATH": "/does/not/exist.json",
		}),
	))
	pre, err := actions.PullRequestEvent(action)
	it.Nil(pre)
	expected := "GitHub Action not triggered for a pull request; Event Name: push"
	it.Equal(expected, fmt.Sprintf("%v", err))

	// 2. Event path does not exist
	tmpDir := tempDir(it)
	eventPath := filepath.Join(tmpDir, "event.json")
	action = githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": eventPath,
		}),
	))
	pre, err = actions.PullRequestEvent(action)
	it.Nil(pre)
	expected = fmt.Sprintf("GitHub Event file does not exist; Path: %s", eventPath)
	it.Equal(expected, fmt.Sprintf("%v", err))

	// 3. General error reading event path
	tmpDir = tempDir(it)
	action = githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": tmpDir,
		}),
	))
	pre, err = actions.PullRequestEvent(action)
	it.Nil(pre)
	expected = fmt.Sprintf("Could not read GitHub Event file; Path: %[1]s\nread %[1]s: is a directory\nis a directory", tmpDir)
	it.Equal(expected, fmt.Sprintf("%v", err))

	// 4. Error parsing event path
	tmpDir = tempDir(it)
	eventPath = filepath.Join(tmpDir, "event.json")
	err = os.WriteFile(eventPath, []byte(`{`), 0644)
	it.Nil(err)
	action = githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": eventPath,
		}),
	))
	pre, err = actions.PullRequestEvent(action)
	it.Nil(pre)
	expected = fmt.Sprintf("Failed to parse GitHub Event file as JSON; Path: %s\nunexpected end of JSON input", eventPath)
	it.Equal(expected, fmt.Sprintf("%v", err))

	// 5. Happy Path
	eventPath = filepath.Join("testdata", "event.json")
	action = githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_EVENT_PATH": eventPath,
		}),
	))
	pre, err = actions.PullRequestEvent(action)
	it.Nil(err)
	// NOTE: The `PullRequestEvent` is WAY too big to make one via a struct
	//       literal so we treat the original file as a golden file.
	asJSON, err := json.MarshalIndent(pre, "", "    ")
	it.Nil(err)
	asJSON = append(asJSON, '\n')
	goldenFilePath := filepath.Join("testdata", "golden.event.json")
	expectedJSON, err := os.ReadFile(goldenFilePath)
	it.Nil(err)
	it.True(bytes.Equal(expectedJSON, asJSON), fmt.Sprintf("Golden File %q does not match expected", goldenFilePath))
}

func TestEventName(t *testing.T) {
	t.Parallel()
	it := assert.New(t)

	action := githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
		}),
	))
	en := actions.EventName(action)
	it.Equal("pull_request", en)
}

func TestRepository(t *testing.T) {
	t.Parallel()
	it := assert.New(t)

	action := githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"GITHUB_REPOSITORY": "blend/certain",
		}),
	))
	r := actions.Repository(action)
	it.Equal("blend/certain", r)
}

func TestRootURL(t *testing.T) {
	t.Parallel()
	it := assert.New(t)

	action := githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"GITHUB_API_URL": "https://ghe.k8s.invalid/api/v3",
		}),
	))
	ru := actions.RootURL(action)
	it.Equal("https://ghe.k8s.invalid/api/v3", ru)
}

func tempDir(it *assert.Assertions) string {
	dir, err := os.MkdirTemp("", "")
	it.Nil(err)

	it.T.Cleanup(func() {
		err := os.RemoveAll(dir)
		it.Nil(err)
	})

	return dir
}
