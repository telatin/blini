# Test Data Generation

This directory contains a Go script for generating test datasets from a large FASTA file of reference genomes. 
The script is used to create datasets for evaluating clustering and searching performance.

## Files in this directory

- `gentestdata.go`: The Go script that generates the data.

## Generated files

The script creates a `testdata/fasta/` directory and populates it with the following files:

- `vir_all.fa`: A FASTA file containing all sampled viral genomes.
- `mut_all.fa`: A FASTA file containing mutated versions (1% SNP) of the sampled genomes.
- `clust_frag.fa`: A FASTA file for testing clustering of fragments. It contains the original genomes and random fragments from each.
- `clust_snps.fa`: A FASTA file for testing clustering with small variations. It contains the original genomes and mutated versions of each.
- `vir_*.fa`: Individual FASTA files for each of the sampled genomes.
- `mut_*.fa`: Individual FASTA files for each of the mutated genomes.

## How to run

1.  Obtain a large reference FASTA file (e.g., from NCBI).
2.  Run the script with the reference file as an argument:
    ```sh
    go run ./paper/gentestdata/gentestdata.go /path/to/your/reference.fa
    ```
3.  The generated test data will be placed in the `testdata/fasta/` directory (relative to the project root).
