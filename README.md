# L4 Load Balancer

This is an implementation of a layer 4 tcp load balancer for educational purposes only.

## Description
TCP server listens at a given address and forwards tcp streams to healthy upstream servers http/https servers.

Monitors health of upstream servers and removes unhealthy servers from rotation and returns it when server is healthly again.

Requests distributed using a round robin algorithm.

## Dependencies
Requires Go 1.22.4
There are no external dependencies.

## Usage
See `./example/main.go` for example.