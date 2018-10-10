# KV-STORE

Implement a set of two apps - node and client.

Node application should act as a storage server which is able to store and retrieve string key-value pairs. In case of multiple nodes launched all nodes should connect to each other and share all keys and values they have including newly added keys on any of the nodes. Nodes should be able to bootstrap themselves with a single IP:port combination of any node already running. Clients should be able to connect to any node with IP:port and key specified and retrieve value for the corresponding key.

## Installation

``` bash
go get -u github.com/spf13/cobra/cobra
```

## Usage

### Server

``` bash
go run main.go server --port 3000 &
go run main.go server --port 4000 &
```

### Client

``` bash
go run main.go client 127.0.0.1:3000 --set a=abcd
go run main.go client 127.0.0.1:3000 --get a
```
