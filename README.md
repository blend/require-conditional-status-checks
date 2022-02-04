# GitHub Action: Require Conditional Status Checks

## Example Usage

```yaml
---
name: 'Meta Workflow: Require Conditional Status Checks'

on:
  pull_request:
    branches:
    - main

jobs:
  meta:
    runs-on:
    - ubuntu-20.04

    steps:
    - name: Ensure All Conditional Checks Have Passed
      uses: blend/require-conditional-status-checks@2022.02.04
      with:
        interval: 20s
        checks-yaml: |
          - job: unit-test-go-core
            paths:
            - cmd/**
            - pkg/**
          - job: lint-go
          - job: protobuf-check-generated
            paths:
            - proto/**
            - pkg/protogen/**
          - job: lint-protobuf
            paths:
            - proto/**
```

Alternatively, the `checks-yaml` can be checked into a file in your repository

```yaml
# ...
    steps:
    - name: Ensure All Conditional Checks Have Passed
      uses: blend/require-conditional-status-checks@2022.02.04
      with:
        interval: 20s
        checks-filename: .github/monorepo/required-checks.yml
# ...
```

## See It In Action

From a recent workflow run on public GitHub:

![Example Workflow Public][2]

From a recent workflow on GitHub Enterprise:

![Example Workflow GHE][3]

## Limitations

-   The GitHub `CompareCommits()` [API][1] can return at most 300 files
    when comparing two commits. This makes it impossible to determine the
    full list of impacted files for PRs with 300 or more files. The list
    of files is critical for `blend/require-conditional-status-checks` to determine which
    checks to enforce.

## Development

This GitHub Action is developed in the `development` [branch][5] of this
repository, in particular via the command in `cmd/requireconditional/`.

The documentation and source code for this action are maintained there. This
branch is intended to be as small as possible so it can be loaded quickly
when new jobs retrieve it when spawned as part of a GitHub Actions workflow.
For more information on how this GitHub Action is developed, see the
[DEVELOPMENT][4] document.

[1]: https://docs.github.com/en/free-pro-team@latest/rest/reference/repos/#compare-two-commits
[2]: https://github.com/blend/require-conditional-status-checks/blob/070363af576e7c117f8f331c6cc11c53c047873a/_images/example-run-public.png?raw=true
[3]: https://github.com/blend/require-conditional-status-checks/blob/070363af576e7c117f8f331c6cc11c53c047873a/_images/example-run-ghe.png?raw=true
[4]: https://github.com/blend/require-conditional-status-checks/blob/070363af576e7c117f8f331c6cc11c53c047873a/DEVELOPMENT.md
[5]: https://github.com/blend/require-conditional-status-checks/tree/070363af576e7c117f8f331c6cc11c53c047873a
