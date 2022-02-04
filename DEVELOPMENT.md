# Development on GitHub Action: Require Conditional Status Checks

## Unit Testing and Code Structure

-   The `cmd/requireconditional/main.go` script is intended to be as short as
    possible so we can maximize the amount of code that can be tested:
    ```go
    package main

    import (
    	githubactions "github.com/sethvargo/go-githubactions"

    	"github.com/blend/require-conditional-status-checks/pkg/requireconditional"
    )

    func main() {
    	action := githubactions.New()
    	err := requireconditional.Run(action)
    	if err != nil {
    		action.Fatalf("%v", err)
    	}
    }
    ```
-   Any code that deals with GitHub Actions inputs, outputs or log
    messages will need to import `github.com/sethvargo/go-githubactions`.
-   When writing (testable) code that uses the `githubactions` package, a
    pointer `action *githubactions.Action` should be used rather than the
    global wrappers around the package `defaultAction`. For example use
    `action.GetInput("role")` instead of `githubactions.GetInput("role")`.
-   In order to test code involving GitHub Actions, it's critical to be able
    to both control environment variables (these are inputs) and to monitor
    writes to STDOUT. In order to do this in tests, both the STDOUT writer and
    the `Getenv()` provider can be replaced:
    ```go
    actionLog := bytes.NewBuffer(nil)
    envMap := map[string]string{
    	"INPUT_ROLE":  "user",
    	"INPUT_LEASE": "3600",
    }
    action := githubactions.New(
    	githubactions.WithWriter(actionLog),
    	githubactions.WithGetenv(func(key string) string {
    		return envMap[key]
    	}),
    )
    // ...
    it.Equal("...", actionLog.String())
    ```

## Testing: Invoking Actions Locally

In order to sanity check an implementation, it can be quite useful to run
an action **locally** instead of doing a pre-release and waiting on a fully
triggered GitHub Actions workflow. To run an action locally, it's enough to
run the `cmd/requireconditional/main.go` binary with the correct environment
variables.

There are two types of environment variables needed; the `GITHUB_*` environment
variables that come with the workflow. See the `dhermes/actions-playground`
[repository][1] for a collection of captured `GITHUB_*` environment variables
from **real** GitHub Actions workflows, organized by event type (e.g.
`pull_request` or `push` event). The other type of inputs are provided in
`inputs:` to the action (i.e. the inputs from `action.yml`). These get
transformed into `INPUT_*` environment variables by GitHub.

For example:

```
env \
  'GITHUB_API_URL=https://api.github.com' \
  'GITHUB_REPOSITORY=dhermes/actions-playground' \
  'GITHUB_EVENT_NAME=pull_request' \
  'GITHUB_EVENT_PATH=./tmp/event.json' \
  "INPUT_GITHUB-TOKEN=$(cat ./tmp/token.txt)" \
  'INPUT_TIMEOUT=30m' \
  'INPUT_INTERVAL=30s' \
  'INPUT_CHECKS-FILENAME=./tmp/checks.yml' \
  go run ./cmd/requireconditional/main.go
```

To locally run an **existing** release of this action, the command would be
similar

```
git clone git@github.com:blend/require-conditional-status-checks.git
cd ./require-conditional-status-checks
env \
  'GITHUB_API_URL=https://api.github.com' \
  'GITHUB_REPOSITORY=dhermes/actions-playground' \
  'GITHUB_EVENT_NAME=pull_request' \
  'GITHUB_EVENT_PATH=./tmp/event.json' \
  "INPUT_GITHUB-TOKEN=$(cat ./tmp/token.txt)" \
  'INPUT_TIMEOUT=30m' \
  'INPUT_INTERVAL=30s' \
  'INPUT_CHECKS-FILENAME=./tmp/checks.yml' \
  node index.js  # Or `./main-darwin-amd64-5786bbfeff4524101e761a8cd8a7d1d05de82d01` directly
```

## Release

-   The action can be "released" by building static binaries and checking them
    in to the `main` branch.
-   The convention is to build binaries (via `make release`) of the form
    `main-${GOOS}-${GOARCH}-${VERSION}` (where `${VERSION}` can be a
    human-readable string or a `git` SHA).
-   Additionally, an `index.js` file should be updated to reflect the
    `${VERSION}` the binaries were built with ([more][2] on why we use
    `index.js` to dispatch our binaries). This can be generated via
    `make generate-index`.
-   The `action.yml` file describing the inputs and outputs and the entrypoint
    binary will need to be copied over to the `main` branch. This file should
    always stay in sync with the code in `pkg/requireconditional/config.go`.
-   The content in the `main` branch can remain minimal and the `README.md` can
    just point back to the `development` branch. By keeping the content
    minimal, running a new workflow via GitHub Actions will be able to download
    the action more quickly.

### Release Helper: `Makefile`

```
$ make  # OR: make help
Makefile for Require Conditional Status Checks GitHub Action

Usage:
   make generate-index        Generate `index.js` file for current VERSION
   make main-darwin-amd64     Build static binary for darwin/amd64
   make main-darwin-arm64     Build static binary for darwin/arm64
   make main-linux-amd64      Build static binary for linux/amd64
   make main-linux-arm64      Build static binary for linux/arm64
   make main-windows-amd64    Build static binary for windows/amd64
   make main-windows-arm64    Build static binary for windows/arm64
   make release               Build all static binaries and `index.js`

```

## FAQ

### Why Not `using: docker`?

By running a static binary **directly** on the host, we don't need to eat
extra costs of `docker pull` or `docker build`.

### Why Not `using: composite`?

We are not able to use `using: composite`. Instead we use `using: node12` and
introduce an `index.js` shim to exec out to the correct static Go binary for
the current `GOOS` and `GOARCH`.

This is due to a limitation of GitHub Actions encountered when using
[actions-runner-controller][5] in GitHub Enterprise (GHE) 3.1 and earlier. The
processes spawned with `using: composite` and `using: node12` had nearly
identical environment variables, but they differed slightly in critical ways.
Environment variables only present for `using: node12` were:

```
ACTIONS_RUNTIME_TOKEN=***
ACTIONS_RUNTIME_URL=https://acme.github.enterprise.invalid/_services/pipelines/...
INPUT_WIDGET=testing
# acme.github.enterprise.invalid is a stand in for the GHE hostname
```

and those only present for `using: composite` were:

```
GITHUB_ACTION_PATH=/runner/_work/_actions/actions/proof-of-concept/main
```

### Are the Release Binaries Optimized?

The `main-${GOOS}-${GOARCH}-${VERSION}` make targets build Go binaries with
custom `-ldflags` to reduce binary size and then use the `upx` [tool][3] to
reduce even further. (See related [blog post][4].)

[1]: https://github.com/dhermes/actions-playground/tree/2021.11.18
[2]: #faq
[3]: https://upx.github.io/
[4]: https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
[5]: https://github.com/actions-runner-controller/actions-runner-controller
