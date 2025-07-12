#! /bin/bash
set -euo pipefail

dir="$(dirname $0)"

set -x
nats stream add schwabn-v0 --config "$dir/schwabn-v0.json"
nats consumer add schwabn-v0 schwabn-futures --config "$dir/schwabn-futures.json"