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

package github

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/blend/go-sdk/ex"
	"github.com/google/go-github/v42/github"
	"golang.org/x/oauth2"
)

// NewClient creates a new client and determines if it's needed for the public
// GitHub API or GitHub Enterprise. The `rootURL` is expected to be the value
// of the `GITHUB_API_URL` environment variable / the `${{ github.api_url }}`
// context value.
func NewClient(ctx context.Context, rootURL, token string) (*github.Client, error) {
	sts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, sts)
	if rootURL == "https://api.github.com" {
		return github.NewClient(tc), nil
	}

	// Root URL for GitHub Enterprise is expected to be `https://[hostname]/api/v3`.
	u, err := url.Parse(rootURL)
	if err != nil {
		return nil, ex.New("GitHub Enterprise root not a valid URL", ex.OptInner(err), ex.OptMessagef("URL: %s", rootURL))
	}
	if u.Path != "/api/v3" {
		return nil, ex.New("GitHub Enterprise root has unexpected path", ex.OptMessagef("URL: %s", rootURL))
	}

	u.Path = "/api/uploads/"
	uploadURL := u.String()
	return github.NewEnterpriseClient(rootURL, uploadURL, tc)
}

// GetFile downloads the contents of a file.
func GetFile(ctx context.Context, c *github.Client, owner, repo, ref, path string) ([]byte, error) {
	opts := &github.RepositoryContentGetOptions{Ref: ref}
	rc, response, err := c.Repositories.DownloadContents(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, ex.New("Failed to download file", ex.OptMessagef("Repository: %s/%s, Ref: %s, Path: %s", owner, repo, ref, path), ex.OptInner(err))
	}
	defer rc.Close()

	if response.StatusCode != http.StatusOK {
		return nil, ex.New("Raw download HTTP failure", ex.OptMessagef("Status Code: %d, Repository: %s/%s, Ref: %s, Path: %s", response.StatusCode, owner, repo, ref, path), ex.OptInner(err))
	}

	body, err := io.ReadAll(rc)
	if err != nil {
		return nil, ex.New("Failed to read body of raw download", ex.OptMessagef("Repository: %s/%s, Ref: %s, Path: %s", owner, repo, ref, path), ex.OptInner(err))
	}

	return body, nil
}
