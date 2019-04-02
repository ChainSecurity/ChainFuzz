#!/bin/bash

set -e

RED='\033[0;31m'
NC='\033[0m'
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'

display_usage() {
	printf "${RED}You must provide path to truffle project that you want to deploy and extract transactions from${NC}\n"
	printf "Usage:\n"
	printf "\t-p, --path\tpath to the first version of truffle project\n"
	printf "\n\n"
	printf "Example:\n\t./extract.sh -p /home/anodar/Desktop/thesis/MetaCoin\n"
	printf "\n\n"
}

if [  $# -le 1 ]
then
	display_usage
	exit 1
fi


POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -p|--path)
    TRUFFLE_DIR="$2"
    shift # past argument
    shift # past value
    ;;
    -s|--solc)
    SOLC_V="$2"
    shift # past argument
    shift # past value
    ;;
    --default)
    DEFAULT=YES
    shift # past argument
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

echo truffle directory        = "${TRUFFLE_DIR}"
VERSION=$(echo ${TRUFFLE_DIR} | sha1sum  | awk '{print $1}')
echo truffle project version  = "${VERSION}"

kill_ganache_if_running() {
	# kill ganache-cli if running (kills node TODO: somehow kill only ganache)
	if pgrep -f /usr/bin/ganache-cli
	then
		pkill -9 -f /usr/bin/ganache-cli
	fi
}

install_solc() {
	# set specified solc version for truffle if present
	if [ ! -z "$SOLC_V" ]
	then
		echo "!!================================================================================!!"
		echo "Please do not use -s with truffle 5, but set the correct solc version in truffle.js"
		echo "!!================================================================================!!"
		exit
	fi
}

extract_transactions() {
	extract_script=$PWD"/build/extract.js"

	# move to truffle directory and deploy contracts
	cd ${TRUFFLE_DIR}
	# convert truffle dir to absolute path
	TRUFFLE_DIR=$PWD
	printf "\n${YELLOW} deploying contracts on Ganache${NC}\n"
	rm -rf build
	truffle deploy # > /dev/null TODO
	printf "\n${GREEN} contracts has been deployed on Ganache${NC}\n"
	# execute transaction extraction javascript in truffle console
	truffle compile
	printf "\n${YELLOW} extracting transactions ${NC}\n"
	truffle exec ${extract_script}
	printf "\n${GREEN} transactions has been extracted${NC}\n"
	# rename transactions file according to project VERSION
	mv ${TRUFFLE_DIR}/transactions.json ${TRUFFLE_DIR}/fuzz_config/transactions_$VERSION.json
}

# check if truffle directories exist
for i in "${TRUFFLE_DIR}"
do
	if [ ! -d $i ]; then
	  printf "\n${RED}Truffle directory: ${NC}$i ${RED} doesn't exist ${NC}\n"
	  exit 1
	fi
done

mkdir -p build/gen
mkdir -p ${TRUFFLE_DIR}/fuzz_config
# write metadata in file
metadata_JSON=${TRUFFLE_DIR}/fuzz_config/metadata_$VERSION.json
metadata="{\n\t\"tuffleProjDir\": \"${TRUFFLE_DIR}\",
\t\"transactions\": \"${TRUFFLE_DIR}/fuzz_config/transactions_${VERSION}.json\",
\t\"accounts\": \"${TRUFFLE_DIR}/fuzz_config/accounts.json\",
\t\"config\": \"${TRUFFLE_DIR}/fuzz_config/config.json\"
}"
printf "$metadata" > $metadata_JSON

# Create default configs if they don't exist
if [ ! -f "${TRUFFLE_DIR}/fuzz_config/config.json" ]; then
    cp ${PWD}/config/config.json  ${TRUFFLE_DIR}/fuzz_config/
fi

if [ ! -f "${TRUFFLE_DIR}/fuzz_config/accounts.json" ]; then
    cp ${PWD}/config/accounts.json  ${TRUFFLE_DIR}/fuzz_config/
fi

# remove transactions json files if exists
if [  -f "${TRUFFLE_DIR}/fuzz_config/transactions_${VERSION}.json" ]; then
	rm "${TRUFFLE_DIR}/fuzz_config/transactions_${VERSION}.json"
fi

kill_ganache_if_running

# install specified version of solc
install_solc

go_project_dir=$PWD

# run ganache-cli on background
python3 $PWD"/build/ganache.py" "${TRUFFLE_DIR}/fuzz_config/accounts.json" > /dev/null &

extract_transactions

kill_ganache_if_running

printf "\t${CYAN}Project metadata was written to: ${GREEN}${metadata_JSON} ${NC}\n"
sleep .5
printf "\t${CYAN} Done. Time to run: ${GREEN} ./build/bin/fuzzer --metadata ${metadata_JSON} --limit 10000 ${NC}\n"
