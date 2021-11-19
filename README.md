# GitHub Action: Composite

## Example Usage

```yaml
---
name: All Checks

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
      uses: blend/action-composite@2021.11.18
      with:
        checks-yaml: |
          - job: sleep1
          - job: sleep2
            paths:
            - changed/**
            - prefix/**
```

Alternatively, the `checks-yaml` can be checked into a file in your repository

```yaml
# ...
    steps:
    - name: Ensure All Conditional Checks Have Passed
      uses: blend/action-composite@2021.11.18
      with:
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
    of files is critical for `blend/action-composite` to determine which
    checks to enforce.

## Development

For more information on how this GitHub Action is developed, see the
[DEVELOPMENT][4] document.

[1]: https://docs.github.com/en/free-pro-team@latest/rest/reference/repos/#compare-two-commits
[2]: _images/example-run-public.png
[3]: _images/example-run-ghe.png
[4]: DEVELOPMENT.md
