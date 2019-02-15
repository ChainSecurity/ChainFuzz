# ChainFuzz: fast transaction fuzzer for Ethereum smart contracts

## Requirements

- ChainFuzz requires a truffle project with correct migration files to fuzz a project.

## Checking custom properties

Functions starting with `fuzz_always_true` will be evaluated for property violations.

If such a function returns a value different than `1` or `True`, then ChainFuzz reports it as a custom property violation and stops.

## Fuzzing Truffle projects using ChainFuzz

The easiest way to use ChainFuzz is using [docker](https://www.docker.com/).

### Build docker image

To build a docker image, run the following command from the root folder of this repository:

```docker build -t chainfuzz .```

### Run docker image

Go to the folder of the truffle project that you would like to test. For testing purposes, you can use `IndividuallyCappedCrowdsale` project:

```cd benchmarks/IndividuallyCappedCrowdsale```

Then, run the following command:

```docker run -v $PWD:/shared -it chainfuzz```

This command starts the docker image in interactive mode and mounts the folder of the truffle project (returned by `$(PWD)`) under `/shared` inside the docker container.

### Set up the fuzzer inside docker

The command above places you inside a new docker container that can run ChainFuzz. Before fuzzing the project, we need to run the following command to deploy the truffle project on ganache and collect fuzzing metadata (described below):

```./build/extract.sh -p /shared/```

### ChainFuzz configuration

ChainFuzz is configured with contracts and functions to be called when fuzzing, as well as which addresses to be used for sending transactions. This configuration is generated automatically by ChainFuzz using the command above, and can be modified by editing the following files (located in folder `fuzz_config`):

- `config.json`: configure which functions should be used / ignored when fuzzing. Example configuration file:
```
{
    "ContractName": {
      "ignore": ["method", "method1"],
      "timestamps": [1503756000, 1803756000]
    },
    "SomeToken": {
      "ignore": ["name", "symbol", "decimals", "pause", "unpause", "renounceOwnership", "transferOwnership"]
    },
    "Migrations": {
      "ignore_all": true
    }
}
```
You can define the functions of any contract to be ignored while fuzzing by mentioning the function name in the `ignore` property. Property `ignore_all` results in ignoring all functions of a given contract. The `timestamps` property takes different timestamps which will be used by the fuzzer to mock the `block.timestamp` and `now` instructions.
- `accounts.json`: Ethereum accounts (addresses with ether balance) that can be used to send the generated transactions


### Run ChainFuzz

To fuzz the project, you need to run ChainFuzz and provide the metadata file generated with the steps above:

```./build/bin/fuzzer --metadata /shared/fuzz_config/metadata_*.json --limit 4000```

Additional options:
- `-o 8` generates additional statistics about the called functions and their failure rates
- `--loglevel=4` provides additional insides into inputs and outputs of the functions

### Results

ChainFuzz checks and reports the following properties:

- Custom property violations (defined as a function whose name starts with `fuzz_always_true` returning something other than 1)
- Violated assertions (these are inserted either implicitly by the solidity compiler or explicitly in the code)
- Arithmetic under-/overflows

For any discovered violation, ChainFuzz generates a JSON file that contains the sequence of transactions that violates the property. 

# Contributors

- Nodar Ambroladze
- [Dr. Hubert Ritzdorf](https://github.com/ritzdorf)
- Petar Ivanov

# License

Licensed under [GNU AFFERO GENERAL PUBLIC LICENSE Version 3](https://www.gnu.org/licenses/agpl-3.0.en.html)

Copyright (C) 2019 [ChainSecurity AG](https://chainsecurity.com)
