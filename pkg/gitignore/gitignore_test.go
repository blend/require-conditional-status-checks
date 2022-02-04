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

package gitignore_test

import (
	"testing"

	"github.com/blend/go-sdk/assert"

	github "github.com/blend/require-conditional-status-checks/pkg/gitignore"
)

func TestGitignoreMatch(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name     string
		Path     string
		Pattern  string
		Expected bool
	}
	cases := []testCase{
		{Name: "Exact Match Yes", Path: "path/to/bob.json", Pattern: "path/to/bob.json", Expected: true},
		{Name: "Exact Match No", Path: "path/to/bill.json", Pattern: "path/to/bob.json", Expected: false},
		{Name: "Ends with ** Yes", Path: "path/to/kerry.json", Pattern: "path/to/**", Expected: true},
		{Name: "Ends with ** No", Path: "path/to/linda.json", Pattern: "path/not/**", Expected: false},
		{Name: "Empty string", Path: "anything.txt", Pattern: "", Expected: false},
		{Name: "Comment", Path: "anything.txt", Pattern: "# Next section", Expected: false},
		{Name: "With Root Path Yes", Path: "a/b/c.json", Pattern: "/a/b/c.json", Expected: true},
		{Name: "With Root Path No", Path: "a/b/d.json", Pattern: "/a/b/c.json", Expected: false},
		{Name: "With Root Path Almost, Absolute/Relative Confusion", Path: "/a/b/c.json", Pattern: "/a/b/c.json", Expected: false},
		{Name: "Trim trailing, Yes", Path: "a/b/c.json", Pattern: "/a/b/c.json\t ", Expected: true},
		{Name: "Trim trailing, No", Path: "a/b/d.json", Pattern: "/a/b/c.json\t ", Expected: false},
		{Name: "Negation, Yes", Path: "b/a/c.json", Pattern: "!a/b/", Expected: true},
		{Name: "Negation, No", Path: "a/b/c.json", Pattern: "!a/b/", Expected: false},
		{Name: "Anywhere in tree, file", Path: "so/many/layers/any.txt", Pattern: "any.txt", Expected: true},
		{Name: "Anywhere in tree, dir", Path: "so/many/layers/here/", Pattern: "here/", Expected: true},
		{Name: "Question mark", Path: "a/b/cXd.txt", Pattern: "a/b/c?d.txt", Expected: true},
		{Name: "Glob, Yes", Path: "a/b/cXYZ123d.txt", Pattern: "a/b/c*d.txt", Expected: true},
		{Name: "Glob, No", Path: "a/b/c/XYZ123/d.txt", Pattern: "a/b/c*d.txt", Expected: false},
		{Name: "Escape characters, No", Path: "a/b/cXd.txt", Pattern: "a/b/c\\?d.txt", Expected: false},
		{Name: "Escape characters, ?", Path: "a/b/c\\?d.txt", Pattern: "a/b/c\\?d.txt", Expected: true},
		{Name: "Escape characters, *", Path: "a/b/c\\*d.txt", Pattern: "a/b/c\\*d.txt", Expected: true},
		{Name: "Match sequence", Path: "aXb.txt", Pattern: "a[a-zA-Z]b.txt", Expected: true},
		{Name: "Normalize globs, One Path", Path: "ab/cd/fg/e.txt", Pattern: "a**b/c***d/**/e.txt", Expected: true},
		{Name: "Normalize globs, Two Paths", Path: "ab/cd/fg/hi/e.txt", Pattern: "a**b/c***d/**/e.txt", Expected: true},
		{Name: "Normalize globs, Zero Paths", Path: "ab/cd/e.txt", Pattern: "a**b/c***d/**/e.txt", Expected: true},
	}
	for _, tc := range cases {
		tc := tc // Re-bind so goroutine spawned by closure gets local reference (vs. range reference)
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			it := assert.New(t)

			actual := github.GitignoreMatch(tc.Path, tc.Pattern)
			it.Equal(tc.Expected, actual)
		})
	}
}

func TestNormalizeGitignore(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name     string
		Pattern  string
		Expected string
	}
	cases := []testCase{
		{Name: "Empty string", Pattern: "", Expected: ""},
		{Name: "Comment", Pattern: "# Next section", Expected: ""},
		{Name: "Exact", Pattern: "/a/b/c.json", Expected: "/a/b/c.json"},
		{Name: "Exact, trim trailing", Pattern: "/a/b/c.json\t ", Expected: "/a/b/c.json"},
		{Name: "Negation", Pattern: "!a/b/", Expected: "!/a/b/**"},
		{Name: "Prepend slash", Pattern: "a/b/c.json", Expected: "/a/b/c.json"},
		{Name: "Prepend double star slash", Pattern: "any.txt", Expected: "**/any.txt"},
		{Name: "Prepend and append double star", Pattern: "here/", Expected: "**/here/**"},
		{Name: "Normalize globs", Pattern: "a**b/c***d/**/e.txt", Expected: "/a*b/c*d/**/e.txt"},
		{Name: "Non double-glob in double glob context (per spec; git behavior differs)", Pattern: "a**b/c***d/***/e.txt", Expected: "/a*b/c*d/*/e.txt"},
	}
	for _, tc := range cases {
		tc := tc // Re-bind so goroutine spawned by closure gets local reference (vs. range reference)
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			it := assert.New(t)

			actual := github.NormalizeGitignore(tc.Pattern)
			it.Equal(tc.Expected, actual)
		})
	}
}
