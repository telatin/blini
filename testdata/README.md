# Test Data for Continuous Integration

This directory contains data and code for sanity-testing Blini's
final executable.

## Files in this directory

- `gen/` data generation script
- `test/` result test script
- `*.fa.zst` simulated data files, generated with `gen`
- `run_blini.sh` runs blini on the test data files

## How to run

1. (Already done) Generate data files with `go run ./testdata/gen`
2. Run Blini on the test data with `run_blini.sh`.
   Results will appear in `tmp/`.
3. Examine the results, either manually or with the script.
   - `clust.json` each cluster should have `ref#` first,
     and `ref#.#` with the same number underneath it.
     There should be three clusters with four elements in each.
     The order is random.
   - `search.csv` each query should be matched with a reference
     with the same number.
     There should be three references and three queries per reference.
     Similarity should be 97%-99%.
     The order is random.
   - An automated script for testing the results can be run with
     `go run ./testdata/test`.
