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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/blend/action-composite/pkg/template"
)

func run() error {
	ctx := context.Background()

	c := template.Config{}
	cmd := &cobra.Command{
		Use:           "template",
		Short:         "Populate `index.template.js`",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return template.Run(ctx, c)
		},
	}

	cmd.PersistentFlags().StringVar(
		&c.TemplateFilename,
		"template-filename",
		c.TemplateFilename,
		"The path to the file where the template is stored",
	)
	cmd.PersistentFlags().StringVar(
		&c.GeneratedFilename,
		"generated-filename",
		c.GeneratedFilename,
		"The path where the generated file should be stored after populating the template",
	)
	cmd.PersistentFlags().StringVar(
		&c.Version,
		"version",
		c.Version,
		"The version being released",
	)

	required := []string{"template-filename", "generated-filename", "version"}
	for _, name := range required {
		err := cobra.MarkFlagRequired(cmd.PersistentFlags(), name)
		if err != nil {
			return err
		}
	}

	return cmd.Execute()
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
