# Dockerfile for handshake POC testbed

FROM ubuntu:16.04

MAINTAINER mkrueger@rstms.net

RUN \
  apt-get update && \
  apt-get install -y git curl sudo

RUN curl https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz | tar -C /usr/local -zx

ENV PATH="$PATH:/usr/local/go/bin"

RUN useradd -ms /bin/bash -G sudo,adm shakes

RUN echo "shakes ALL=(ALL) NOPASSWD:ALL" >>/etc/sudoers

USER shakes

WORKDIR /home/shakes

RUN mkdir go

ENV GOPATH /home/shakes/go

RUN \ 
  go get github.com/nomasters/hashmap && \
  go get go.etcd.io/bbolt && \
  go get github.com/fatih/color && \ 
  go get github.com/spf13/cobra && \
  go get github.com/spf13/viper && \
  go get gopkg.in/yaml.v2

RUN \
  cd /home/shakes/go/src/github.com/nomasters && \
  git clone https://github.com/nomasters/handshake && \
  cd handshake && \
  git checkout interfaces-draft && \
  cd cmd/handshake && \
  go install 
  
ENV PATH="$PATH:~/go/bin"

ENTRYPOINT /bin/bash
