# Contribution Guidelines

Everyone is welcome to contribute to Blini.

## Reporting bugs

If you encounter a bug, please open an
[issue](https://github.com/fluhus/blini/issues/new/choose)
and report it.
The most important thing it to help us reproduce the bug so we can fix it.

## Requesting new features

Feel free to open an
[issue](https://github.com/fluhus/blini/issues/new/choose)
and describe your new feature request.
Try to explain your use case and what problem the new feature is meant to solve.

## Pull requests

Code contributions are welcome.

**Before you write new code:**
- Discuss the new change in an issue or a discussion.
  Make sure the idea is welcome and that everyone is on the same page
  regarding the problem and the solution.
- If you are solving an obvious bug with an obvious fix,
  you can go ahead without discussing it before.

**Before you create a pull request:**
- Make sure the code is formatted with `goimports` or with `gopls`.
- Make sure the automated integration tests pass.

## Repository structure

- `blini`, `sketching`: source code
- `paper`: publication-related text, plots and scripts
- `testdata`: mock data for integration testing
- `build_release.sh`: builds the release binaries
