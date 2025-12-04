# Runs Blini on the generated test data.

set -e

mkdir -p tmp
blini -q testdata/clust.fa.zst -o tmp/clust -c
blini -r testdata/refs.fa.zst -q testdata/queries.fa.zst -o tmp/search.csv -c

echo "Done. You can find the results in the tmp/ directory."
