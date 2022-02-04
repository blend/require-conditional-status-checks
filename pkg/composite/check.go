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

package composite

import (
	"github.com/blend/go-sdk/ex"
	"github.com/google/go-github/v42/github"

	"github.com/blend/require-conditional-status-checks/pkg/gitignore"
)

const (
	// commitsComparisonMaxFiles is the limit for the number of files that
	// can be returned from the GitHub API. Once the limit is reached, we
	// just give up and return an error.
	commitsComparisonMaxFiles = 300
)

// Check represents a required check on a pull request.
type Check struct {
	Job   string   `json:"job,omitempty" yaml:"job,omitempty"`
	Paths []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

// Required determines if a check should be required for a given pair of
// compared commits.
func (c Check) Required(cc *github.CommitsComparison) (bool, error) {
	// Early exit if there are "too many" files for the comparison.
	if len(cc.Files) >= commitsComparisonMaxFiles {
		err := ex.New("Commit Comparison contained too many files", ex.OptMessagef("File Count: %d", len(cc.Files)))
		return false, err
	}

	if len(c.Paths) == 0 {
		return true, nil
	}

	for _, file := range cc.Files {
		for _, pattern := range c.Paths {
			if gitignore.GitignoreMatch(file.GetFilename(), pattern) {
				return true, nil
			}

			if file.GetStatus() == "renamed" && gitignore.GitignoreMatch(file.GetPreviousFilename(), pattern) {
				return true, nil
			}
		}
	}

	return false, nil
}
