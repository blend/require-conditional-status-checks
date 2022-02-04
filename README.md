# GitHub Action: Composite

## Example Usage

```yaml
---
name: 'Meta Workflow: Composite'

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
      uses: blend/require-conditional-status-checks@2022.01.21
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
      uses: blend/require-conditional-status-checks@2022.01.21
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

For more information on how this GitHub Action is developed, see the
[DEVELOPMENT][4] document.

[1]: https://docs.github.com/en/free-pro-team@latest/rest/reference/repos/#compare-two-commits
[2]: _images/example-run-public.png
[3]: _images/example-run-ghe.png
[4]: DEVELOPMENT.md
