import csv
from collections import defaultdict

import numpy as np
from matplotlib import pyplot as plt
from myplot import ctx


def load_table(f: str):
    d = defaultdict(list)
    for row in csv.DictReader(open(f)):
        for k, v in row.items():
            d[k].append(v)
    for k in d:
        try:
            d[k] = [float(x) for x in d[k]]
        except ValueError:
            continue
    d['time_seconds'] = list(zip(*[d[f'time_seconds_{i}'] for i in range(1, 6)]))
    d['time_seconds'] = [np.mean(x) for x in d['time_seconds']]
    return d


def plot_search():
    d = load_table('results_search.txt')
    print(d)
    d['name'] = [x.replace(' ', '\n') for x in d['name']]
    with ctx('search_time', sizeratio=0.75):
        plt.bar(d['name'][:2], d['time_seconds'][:2])
        plt.bar(d['name'][2:4], d['time_seconds'][2:4])
        plt.bar(d['name'][4:], d['time_seconds'][4:])
        plt.yscale('log')
        plt.ylabel('Average time (s)')
        plt.xticks(rotation=20)
    with ctx('search_found', sizeratio=0.75):
        plt.bar(d['name'][:2], d['source_found'][:2])
        plt.bar(d['name'][2:4], d['source_found'][2:4])
        plt.bar(d['name'][4:], d['source_found'][4:])
        plt.ylabel('Successful matches (out of 100)')
        plt.xticks(rotation=20)
    with ctx('search_others', sizeratio=0.75):
        plt.bar(d['name'][:2], d['others_found'][:2])
        plt.bar(d['name'][2:4], d['others_found'][2:4])
        plt.bar(d['name'][4:], d['others_found'][4:])
        plt.ylabel('Non-source matches ')
        plt.xticks(rotation=20)


def plot_clust(tag: str):
    d = load_table(f'results_clust_{tag}.txt')
    # d['name'] = [x.replace(' (', '\n(') for x in d['name']]
    print(d)
    with ctx(f'clust_{tag}_time', sizeratio=0.75):
        plt.bar(d['name'][:4], d['time_seconds'][:4])
        plt.bar(d['name'][4:], d['time_seconds'][4:])
        plt.ylabel('Average time (s)')
        plt.xticks(rotation=15)
    with ctx(f'clust_{tag}_mem', sizeratio=0.75):
        plt.bar(d['name'][:4], d['max_mem_mb'][:4])
        plt.bar(d['name'][4:], d['max_mem_mb'][4:])
        plt.ylabel('Max memory (MB)')
        plt.xticks(rotation=15)
    with ctx(f'clust_{tag}_ari', sizeratio=0.75):
        plt.bar(d['name'][:4], d['ari'][:4])
        plt.bar(d['name'][4:], d['ari'][4:])
        plt.ylabel('Adjusted Rand-Index')
        plt.xticks(rotation=15)
        ylim = [min(d['ari']), max(d['ari'])]
        stretch = (ylim[1] - ylim[0]) * 0.2
        ylim[0] -= stretch
        ylim[1] += stretch
        plt.ylim(ylim)
    with ctx(f'clust_{tag}_nclust', sizeratio=0.75):
        plt.bar(d['name'][:4], d['n_clusters'][:4])
        plt.bar(d['name'][4:], d['n_clusters'][4:])
        plt.ylabel('Number of clusters')
        plt.xticks(rotation=15)


plt.style.use('ggplot')
plot_search()
# plot_clust('frag')
# plot_clust('snps')
