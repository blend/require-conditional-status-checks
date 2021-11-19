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

package template

import (
	"context"
	"html/template"
	"os"
)

// Run reads a template file, populates it with the configured version and
// writes out the populated template to the generated filename.
func Run(_ context.Context, c Config) error {
	t, err := template.ParseFiles(c.TemplateFilename)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(c.GeneratedFilename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, c)
}
