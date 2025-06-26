# Runs benchmarks on blini, sourmash and mmseqs.

## DATA GENERATION

# Generate simulated dataset.
go run ./publication

# Pick random viral genomes.
go run ./x/refseqpick refseq/viral.1.1.genomic.fna.gz

## SEARCH TASKS

# Blini sketch reference.
./blini -r refseq/viral.1.1.genomic.fna.gz -o viral.blini

# Search genomes with blini.
time ./blini -q publication/testdata/vir_all.fa -r viral.blini -o tmp.blini_vir.csv
time ./blini -q publication/testdata/vir_mut.fa -r viral.blini -o tmp.blini_mut.csv

# Sourmash sketch reference.
sourmash sketch dna --singleton refseq/viral.1.1.genomic.fna.gz -o viral.sm

# Searches a single file with sourmash.
function smsearch {
  sourmash sketch dna --singleton "$1" -o tmp.sm &> /dev/null &&
  sourmash search -q --ignore-abundance -o "$2" tmp.sm viral.sm &&
  rm tmp.sm
}

# Search genomes with sourmash.
time find publication/testdata/ -name 'vir_[0-9]*' |
while read f; do
  smsearch $f tmp.sm_$(basename $f .fa).csv
done

time find publication/testdata/ -name 'mut_[0-9]*' |
while read f; do
  smsearch $f tmp.sm_$(basename $f .fa).csv
done

# Evaluate searches.
go run ./publication/testsearch

## CLUSTER SNPS

# Blini cluster.
for s in 25 50 100 200; do
  time ./blini -q publication/testdata/snps.fa -o tmp.blini_$s -s $s -c
done

# MMseq cluster.
time \
  mmseqs easy-linclust --threads 1 \
  publication/testdata/snps.fa tmp.mm tmp --min-seq-id 0.9 --cov-mode 1

# Evaluate clustering.
go run ./publication/testclust

## CLUSTER FRAGMENTS

# Blini cluster.
for s in 25 50 100 200; do
  time ./blini -q publication/testdata/frag.fa -o tmp.blini_$s -s $s -c
done

# MMseq cluster.
time \
  mmseqs easy-linclust --threads 1 \
  publication/testdata/frag.fa tmp.mm tmp --min-seq-id 0.9 --cov-mode 1

# Evaluate clustering.
go run ./publication/testclust
