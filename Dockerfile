FROM ubuntu:18.04

RUN apt-get update \
  && apt-get install -y software-properties-common \
  && add-apt-repository -y ppa:ethereum/ethereum \
  && apt-get update

# install golang 1.10
RUN apt-get install -y golang-1.10-go
# export GOPATH
RUN printf "export GOPATH=/go\n" >> ~/.bashrc \
  && printf "export PATH=$PATH:/usr/lib/go-1.10/bin\n" >> ~/.bashrc \
  && printf "echo Welcome to the FUZZER\n" >> ~/.bashrc \
  && printf "echo Important commands are in bash history.\n" >> ~/.bashrc

# use bash instead of /bin/sh
RUN ["/bin/bash", "-c", "source ~/.bashrc"]

RUN apt-get update && apt-get install -y git vim

RUN echo './build/extract.sh -p /shared/ \n./build/bin/fuzzer --metadata /shared/fuzz_config/metadata_*.json --limit 4000 -o 8 --loglevel=4' > /root/.bash_history


# install nodejs, Truffle and Ganache
RUN apt-get install -y curl \
  && curl -sL https://deb.nodesource.com/setup_11.x |  bash - 

RUN apt-get install -y nodejs \
  && npm -g config set user root \
  && npm install -g ganache-cli \
  && npm install -g truffle

# clone go-ethereum, reset to revision: 27e3f968194e2723279b60f71c79d4da9fc7577f 
# and apply patch
COPY ./build/patch /tmp/patch
RUN mkdir -p /go/src/github.com/ethereum && cd /go/src/github.com/ethereum \
  && git clone https://github.com/ethereum/go-ethereum.git \
  && cd /go/src/github.com/ethereum/go-ethereum \
  && git reset --hard 27e3f968194e2723279b60f71c79d4da9fc7577f \
  && cd /go/src/github.com/ethereum/go-ethereum && git apply /tmp/patch

WORKDIR /go/src/fuzzer

COPY . /go/src/fuzzer/

RUN ["/bin/bash", "-ci", "make fmt; make fuzz > /dev/null"]

