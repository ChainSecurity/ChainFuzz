.PHONY: all

export GOBIN = $(shell pwd)/build/bin

fuzz:
	build/env.sh go install fuzzer/*.go

# extract transactions from truffle deployment code
tx:
	@build/extract.sh ${version} ${path} ${solc}

fmt:
	go fmt utils/*
	go fmt fuzzer/*
	go fmt argpool/*
