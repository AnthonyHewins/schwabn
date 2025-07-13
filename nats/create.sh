#! /bin/bash
set -euo pipefail

dir="$(dirname $0)"

set -x
nats stream add schwabn-v0 --config "$dir/schwabn-v0.json"

for i in schwabn-futures schwabn-chart-futures; do
    nats consumer add schwabn-v0 $i --config "$dir/$i.json"
done
