from __future__ import print_function
import argparse
import ip
import kv_pb2
import kv_pb2_grpc
import grpc
import logging
import re
import sys
import time

import ip
import kv_pb2
import kv_pb2_grpc

args = {}

class ConfigError(Exception):
    pass

def setCustomLogger(name):
    formatter = logging.Formatter(fmt="%(asctime)s: %(message)s", datefmt='%Y-%m-%d %H:%M:%S')
    screen_handler = logging.StreamHandler(stream=sys.stdout)
    screen_handler.setFormatter(formatter)
    logger = logging.getLogger(name)
    logger.setLevel(logging.INFO)
    logger.addHandler(screen_handler)
    return logger

logger = setCustomLogger("KV")

def explain(msg):
  if args.verbose:
    logger.info("%s" % msg)

def parseArgs():
    global args
    parser = argparse.ArgumentParser(description="Send request(s) to a key-value store server.")
    parser.add_argument("--ip", "-i", nargs=1, action="store",
                        help="set server IP address (IPv4 only!).")
    parser.add_argument("--get", "-g", nargs=1, action="append",
                        help="get value associated with key.")
    parser.add_argument("--set", "-s", nargs=1, action="append",
                        help="set a key-value pair.")
    parser.add_argument("--list", "-l", action="store_true",
                        help="get key-value pairs defined on server.")
    parser.add_argument("--verbose", "-v", action="store_true",
                        help="verbosely list operations performed.")
    args = parser.parse_args()

    if args.ip == None:
        args.ip = ["127.0.0.1:4000"]
    if ip.isValidIP(args.ip[0]) == False:
        raise ConfigError("not a valid server IP address: '%s'" % args.ip[0])

def doGet(ip, key):
    explain("sending GET request to {0:s} for key '{1:s}'...".format(ip[0], key))
    with grpc.insecure_channel(args.ip[0]) as channel:
        stub = kv_pb2_grpc.ClientStub(channel)
        response = stub.Get(kv_pb2.GetKey(key=key))
        if response.defined:
            print("'%s'='%s'" % (key, response.value))
        else:
            print("'%s': undefined" % key)

def doSet(ip, key, value):
    explain("sending SET request to {0:s} for key '{1:s}'...".format(ip[0], key))
    with grpc.insecure_channel(args.ip[0]) as channel:
        stub = kv_pb2_grpc.ClientStub(channel)
        stub.Set(kv_pb2.SetKey(key=key, value=value, broadcast=True))

def doList(ip):
    explain("sending LIST request to %s..." % ip[0])
    with grpc.insecure_channel(args.ip[0]) as channel:
        stub = kv_pb2_grpc.ClientStub(channel)
        response = stub.List(kv_pb2.Void())
        print("Key-value pairs defined on %s:" % ip[0])
        for key in response.store:
            print("  - '{0:s}'='{1:s}'".format(key, response.store[key]))
        print("-- end of key-value dump --")

def handleGet(ip, key):
    regex = re.compile('^[a-zA-Z0-9_]+$')
    if regex.match(key) == False:
        raise ConfigError("invalid --get: expected '--get KEY'; got '--get %s'" % key)
    doGet(ip, key)

def handleSet(ip, kv):
    regex = re.compile('^([a-zA-Z0-9_]+)=([a-zA-Z0-9_]+)$')
    m = regex.match(kv)
    if m == None:
         raise ConfigError("invalid --set: expected '--set KEY=VALUE'; got '--get %s'" % kv)
    doSet(ip, m.group(1), m.group(2))

def run():
    try:
        parseArgs()
        if args.get != None:
            for key in args.get:
                handleGet(args.ip, key[0])
        if args.set != None:
            for kv in args.set:
                handleSet(args.ip, kv[0])
        if args.list:
            doList(args.ip)
    except ConfigError as e:
        print("error:", e.args[0])
    except grpc.RpcError as e:
        print("error: {0:s}: {1:s}".format(e.code(), e.details()))

if __name__ == '__main__':
    run()
