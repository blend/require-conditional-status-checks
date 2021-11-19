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
	"time"

	"github.com/blend/go-sdk/ex"
	githubactions "github.com/sethvargo/go-githubactions"
)

func wrapError(err error, name, value string) error {
	// NOTE: Placing the `value` in the error message may not be desired
	//       for sensitive (but malformed) inputs. GitHub Actions will scrub
	//       any secrets from output, but if the sensitive value is not a
	//       secret available to the workflow, it will not be scrubbed.
	failure := ex.New("Invalid input", ex.OptMessagef("Input: %q, Value: %q", name, value))
	return ex.Nest(failure, err)
}

// DurationInput parses GitHub Actions inputs that are intended to be a `time.Duration`.
func DurationInput(action *githubactions.Action, name string) (time.Duration, error) {
	valueStr := action.GetInput(name)
	d, err := time.ParseDuration(valueStr)
	if err != nil {
		err = wrapError(err, name, valueStr)
	}
	return d, err
}
