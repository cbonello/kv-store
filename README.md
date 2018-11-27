# KV-STORE

A small key-value database to play with [gRPC](https://grpc.io/),
[protocol buffers](https://developers.google.com/protocol-buffers/)
and programming language interoperability.

The database can be replicated on multiple servers that implement a basic [eventual consistency](https://en.wikipedia.org/wiki/Eventual_consistency) model using
the **RegisterWithPeer()** and **Set()** rpc methods (see file [kv.proto](./kv.proto)):

1. When a new server is started, it performs a **RegisterWithPeer()** rpc call for each peer IP listed on the command line. Peers respond with the contents of their key-value store.

2. When a key-value is created/updated, a server broadcast the update to its peer(s).

Current implementation restricts Keys and values to alphanumeric characters and underscores. Both a client and a server are provided and implemented with Google Go, Node.js and Python.

*Repository was only tested on Linux and is using GNU make features.*

## Installation

``` bash
make install
make build
```

## Usage

To demonstrate the servers functionnalities, let's start a few of them.

### Server

#### Go server

``` bash
$ go run go/main.go server -v
Listening on 127.0.0.1:4000...
```

Note: Default IP address and port number used by servers and clients is "127.0.0.1:4000".

#### Node.js server

```bash
$ node node/index.js server --ip 127.0.0.1:4001 -v 127.0.0.1:4000
1543334553169: registering with peer 127.0.0.1:4000...
Listening on 127.0.0.1:4001...
```

#### Python server

```bash
$ python python/server.py --ip 127.0.0.1:4002 -v 127.0.0.1:4000 127.0.0.1:4001
2018-11-27 11:08:08: registering with peer 127.0.0.1:4000...
2018-11-27 11:08:08: registering with peer 127.0.0.1:4001...
Listening on 127.0.0.1:4002...
```

### Client

``` bash
$ python python/client.py --set a=b --list -v
2018-11-27 11:26:42: sending SET request to 127.0.0.1:4000 for key 'a'...
2018-11-27 11:26:42: sending LIST request to 127.0.0.1:4000...
Key-value pairs defined on 127.0.0.1:4000:
  - 'a'='b'
-- end of key-value dump --
```

``` bash
$ go run go/main.go client --ip 127.0.0.1:4001 --list
Key-value pairs defined on 127.0.0.1:4001:
  - 'a'='b'
-- end of key-value dump --
```

``` bash
$ node node/index.js client --ip 127.0.0.1:4002 --list -v
1543335852521: sending LIST request to 127.0.0.1:4002...
Key-value pairs defined on 127.0.0.1:4002:
 - 'a'='b'
-- end of key-value dump --
```

## TODO

1. Add unit tests.
2. Handle network errors.