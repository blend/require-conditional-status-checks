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
	"testing"

	githubactions "github.com/sethvargo/go-githubactions"

	"github.com/blend/go-sdk/assert"
)

func Test_actionGetEnv(t *testing.T) {
	it := assert.New(t)

	calls := []string{}
	getenv := func(k string) string {
		calls = append(calls, k) // NOTE: This is not concurrency safe
		return k + k + k
	}

	action := githubactions.New(githubactions.WithGetenv(getenv))
	it.Equal("foofoofoo", actionGetEnv(action, "foo"))
	it.Equal([]string{"foo"}, calls)
	it.Equal("barbarbar", actionGetEnv(action, "bar"))
	it.Equal([]string{"foo", "bar"}, calls)
}
