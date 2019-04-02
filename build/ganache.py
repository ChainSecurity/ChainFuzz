#!/usr/bin/python3

import json
import subprocess
import sys
from pprint import pprint

print(sys.argv)
if len(sys.argv) < 2:
    print("you must provide accounts file")
    exit(1)

# read the accounts (key, balance) and start ganache-cli with those accounts
cmd = ""
with open(sys.argv[1]) as f:
    data = json.load(f)
    for account in data['accounts']:
        cmd = "{} --account=\"0x{},{}\"".format(cmd, account['key'], account['amount'])

process = subprocess.call(["ganache-cli  --allowUnlimitedContractSize --gasLimit 0xfffffffffff {} > /dev/null".format(cmd)], shell=True)
