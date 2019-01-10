rm -rf output/*
mkdir -p output/coordinator output/worker
jsonnet presto.jsonnet -S -m output
