# Blini

Lightweight nucleotide sequence searching and dereplication.

## Requirements

None.
[Download](https://github.com/fluhus/blini/releases)
and get started!

## Usage

### Searching

With both `-q` and `-r` set, Blini looks up the query entries in the
reference entries.
The reference may either be a fasta or a pre-sketched index.

```sh
blini -q query.fasta -r reference.fasta -o output.csv
# Or
blini -q query.fasta -r reference.blini -o output.csv
```

### Sketching

With only `-r` set, Blini pre-sketches the given reference for use
in search operations.
This makes lookup operations quicker.

```sh
blini -r reference.fasta -o reference.blini
```

### Clustering

With only `-q` set, Blini dereplicates (clusters) the query set.

```sh
blini -q input.fasta -o output_prefix
```

The outputs are a fasta file with the representatives,
and a JSON file with the cluster assignments.

### Other options

* `-h` display help on the available flags.
* `-c` for searching, calculate containment of query in the reference
  rather than full match.
* `-m` for searching and clustering,
  minimal similarity for a match.
* `-s` scale; use 1/s of kmers for similarity.

## Limitations

* Blini supports nucleotide sequences only.
  Amino-acids are currently not supported.
* Blini runs on a single file with sequences,
  where each sequence is a separate species.
  Support for multiple files and multiple sequences per species
  will be added in the future.
* No multi-threading at the moment.
  Still fast, innit?
