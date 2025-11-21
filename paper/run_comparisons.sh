# Runs benchmarks on blini, sourmash and mmseqs.

datadir=
outdir=
plotdir=
cmbfigdir=

## DATA GENERATION

# Generate simulated dataset.
go run ./paper/gentestdata $datadir/viral.1.1.genomic.fna.gz


# PRE-SKETCH REFERENCE DATASET

blini -r $datadir/viral.1.1.genomic.fna.gz -o $outdir/viral.blini
sourmash sketch dna --singleton $datadir/viral.1.1.genomic.fna.gz -o $outdir/viral.sm
mmseqs createdb $datadir/viral.1.1.genomic.fna.gz $outdir/viral.mm


## SEARCH TASKS

# Search genomes with blini.
time blini -q testdata/fasta/vir_all.fa -r $outdir/viral.blini -o $outdir/blini_vir.csv
time blini -q testdata/fasta/mut_all.fa -r $outdir/viral.blini -o $outdir/blini_mut.csv

# Searches a single file with sourmash.
function smsearch {
  sourmash sketch dna --singleton "$1" -o tmp.sm &> /dev/null &&
  sourmash search -q --ignore-abundance -o "$2" tmp.sm viral.sm &&
  rm tmp.sm
}

# Search genomes with sourmash.
time find testdata/fasta/ -name 'vir_[0-9]*' |
while read f; do
  smsearch $f $outdir/sm_$(basename $f .fa).csv
done

time find testdata/fasta/ -name 'mut_[0-9]*' |
while read f; do
  smsearch $f $outdir/sm_$(basename $f .fa).csv
done

# Search genomes with mmseqs.
time mmseqs easy-search --search-type 3 --threads 1 --min-seq-id 0.9 testdata/fasta/vir_all.fa $outdir/viral.mm tmp.mms tmp
time mmseqs easy-search --search-type 3 --threads 1 --min-seq-id 0.9 testdata/fasta/mut_all.fa $outdir/viral.mm tmp.mmsm tmp

# Evaluate searches.
go run ./paper/testsearch


## CLUSTERING

# Blini cluster.
for s in 25 50 100 200; do
  reset
  for i in 1 2 3 4 5; do
    time blini -q testdata/clust_snps.fa -o tmp.blini_$s -s $s -c -m 0.97
  done
  read -p "Done $s"
done

# MMseqs cluster.
for t in 1 4; do
  reset
  for i in 1 2 3 4 5; do
    rm -fr tmp &&
    time \
      mmseqs easy-linclust -v 1 --threads $t \
      testdata/clust_snps.fa tmp.mm tmp \
      --min-seq-id 0.97 --seq-id-mode 1 --cov-mode 1 &&
    rm -fr tmp
  done
  read -p "Done $t"
done

# Evaluate clustering.
go run ./paper/testclust


## PLOT COMBINING

# Search
python $cmbfigdir/cmbfig.py \
  -c 3 -o results/search.png -i \
  $plotdir/search_found.png \
  $plotdir/search_others.png \
  $plotdir/search_time.png

# Cluster frag
python $cmbfigdir/cmbfig.py \
  -c 2 -o results/clust_frag.png -i \
  $plotdir/clust_frag_nclust.png \
  $plotdir/clust_frag_ari.png \
  $plotdir/clust_frag_time.png \
  $plotdir/clust_frag_mem.png

# Cluster snps
python $cmbfigdir/cmbfig.py \
  -c 2 -o results/clust_snps.png -i \
  $plotdir/clust_snps_nclust.png \
  $plotdir/clust_snps_ari.png \
  $plotdir/clust_snps_time.png \
  $plotdir/clust_snps_mem.png
