FROM golang:1.11.5

ENV GO111MODULE=on
WORKDIR /go/src/github.com/nomasters/hashmap

# copy over files important to the project
COPY go.mod .
COPY *.go ./
COPY cmd/ cmd

# install the commandline tools
RUN go install github.com/nomasters/handshake/cmd/handshake
