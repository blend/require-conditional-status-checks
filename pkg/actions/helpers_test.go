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
	"fmt"
	"testing"
	"time"

	"github.com/blend/go-sdk/assert"
	githubactions "github.com/sethvargo/go-githubactions"

	"github.com/blend/require-conditional-status-checks/pkg/actions"
)

func TestDurationInput(t *testing.T) {
	t.Parallel()
	it := assert.New(t)

	action := githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"INPUT_TEST-DURATION": "1s",
		}),
	))
	input, err := actions.DurationInput(action, "test-duration")
	it.Nil(err)
	it.Equal(time.Second, input)

	// Failure
	action = githubactions.New(githubactions.WithGetenv(
		getenvFromMap(map[string]string{
			"INPUT_TEST-DURATION": "x",
		}),
	))
	input, err = actions.DurationInput(action, "test-duration")
	it.Equal(0, input)
	it.NotNil(err)
	it.Equal("Invalid input; Input: \"test-duration\", Value: \"x\"\ntime: invalid duration \"x\"", fmt.Sprintf("%v", err))
}

func getenvFromMap(m map[string]string) githubactions.GetenvFunc {
	return func(key string) string {
		return m[key]
	}
}
