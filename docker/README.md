# Handshake POC Docker testbed

This dockerfile installs and builds the *interfaces-draft* branch of handshake for testing.

It uses the official golang binaries from https://golang.org/dl/ and the installation instructions from https://golang.org/doc/install

 Command      | Description
 ------------ | -----------------------------------------
 `make build` | Create a handshake docker image
 `make shell` | Run an interactive shell in the container
 `make clean` | Delete all docker images and containers
