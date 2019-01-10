run_jsonnet:
	mkdir -p output/coordinator output/worker
	jsonnet presto.jsonnet -S -m output

.PHONY: clean
clean:
	rm -rf output
