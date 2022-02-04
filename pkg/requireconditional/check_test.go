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

package requireconditional_test

import (
	"fmt"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/ref"
	githubofficial "github.com/google/go-github/v42/github"

	requireconditional "github.com/blend/require-conditional-status-checks/pkg/requireconditional"
)

func TestCheckRequired(t *testing.T) {
	t.Parallel()
	it := assert.New(t)

	// Early exit when exceeding limit (this uses more memory than we'd like
	// but that's OK)
	c := requireconditional.Check{Job: "shark", Paths: []string{"week/**"}}
	cc := &githubofficial.CommitsComparison{Files: make([]*githubofficial.CommitFile, 300)}
	required, err := c.Required(cc)
	it.False(required)
	it.Equal("Commit Comparison contained too many files; File Count: 300", fmt.Sprintf("%v", err))

	// Early exit when there are no paths.
	c = requireconditional.Check{Job: "all-dem"}
	cc = &githubofficial.CommitsComparison{}
	required, err = c.Required(cc)
	it.Nil(err)
	it.True(required)

	// No matches
	c = requireconditional.Check{Job: "shark", Paths: []string{"week/**"}}
	cc = &githubofficial.CommitsComparison{Files: []*githubofficial.CommitFile{
		{Filename: ref.String("month/fish.txt")},
	}}
	required, err = c.Required(cc)
	it.Nil(err)
	it.False(required)

	// Match on filename
	c = requireconditional.Check{Job: "shark", Paths: []string{"week/**"}}
	cc = &githubofficial.CommitsComparison{Files: []*githubofficial.CommitFile{
		{Filename: ref.String("week/fish.txt")},
	}}
	required, err = c.Required(cc)
	it.Nil(err)
	it.True(required)

	// Match on `previous_filename`
	c = requireconditional.Check{Job: "shark", Paths: []string{"week/**"}}
	cc = &githubofficial.CommitsComparison{Files: []*githubofficial.CommitFile{
		{
			Status:           ref.String("renamed"),
			Filename:         ref.String("month/fish.txt"),
			PreviousFilename: ref.String("week/fish.txt"),
		},
	}}
	required, err = c.Required(cc)
	it.Nil(err)
	it.True(required)
}
