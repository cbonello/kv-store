const grpc = require('grpc')
const protoLoader = require('@grpc/proto-loader')
const path = require('path')
const isValidIP = require('../ip')

const PROTO_PATH = path.join(__dirname, '/../../../kv.proto')
const packageDefinition = protoLoader.loadSync(
  PROTO_PATH,
  {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
  })
const KVProto = grpc.loadPackageDefinition(packageDefinition).kv

let ip = ''
let verbose = false

const store = {}
const peers = []

const explain = msg => {
  if (verbose) {
    console.log(`${Date.now()}: ${msg}`)
  }
}

const get = (call, callback) => {
  const key = call.request.key
  if (key in store) {
    explain(`received GET request for key '${key}': value = '${store[key]}'`)
    callback(null, { value: store[key], defined: true })
  } else {
    explain(`received GET request for key '${key}': value = undefined`)
    callback(null, { value: store[key], defined: false })
  }
}

const set = (call, callback) => {
  const key = call.request.key
  const value = call.request.value
  const broadcast = call.request.broadcast
  store[key] = value
  if (broadcast) {
    explain(`received SET request for key '${key}': new value = '${value}'`)
    updatePeers(key, value)
  } else {
    explain(`received peer update for key '${key}': new value = '${value}'`)
  }
  callback(null, { value })
}

const list = (call, callback) => {
  explain(`received LIST request`)
  callback(null, { store })
}

const registerWithPeer = (call, callback) => {
  const ip = call.request.ip
  explain(`received new peer registration: ${ip}`)
  if (isValidIP(ip)) {
    if (!peers.includes(ip)) {
      peers.push(ip)
    }
  }
  callback(null, { store })
}

const handleRequests = argv => {
  const server = new grpc.Server()
  server.addService(
    KVProto.Client.service,
    {
      get: get,
      set: set,
      list: list,
      registerWithPeer: registerWithPeer
    }
  )
  server.bind(argv.ip, grpc.ServerCredentials.createInsecure())
  console.log(`Listening on ${argv.ip}...`)
  server.start()
}

const introduceOurself = peerIP => {
  const grpcClient = new KVProto.Client(peerIP, grpc.credentials.createInsecure())
  explain(`registering with peer ${peerIP}...`)
  grpcClient.registerWithPeer({ ip }, function (err, response) {
    if (err) {
      console.error(`error: ${err.message}`)
    } else {
      for (const key in response.store) {
        store[key] = response.store[key]
      }
    }
  })
}

const updatePeers = (key, value) => {
  for (const peerIP in peers) {
    explain(`updating peer '${peers[peerIP]}': '${key}' = '${value}'`)
    const grpcClient = new KVProto.Client(peers[peerIP], grpc.credentials.createInsecure())
    grpcClient.set({ key, value, broadcast: false }, function (err, response) {
      if (err) {
        console.error(`error: ${err.message}`)
      }
    })
  }
}

const server = {
  desc: 'Key-value store server.',
  builder: yargs =>
    yargs
      .positional('peer', {
        describe: `Set peer IP address (IPv4 only!).`,
        type: 'string'
      })
      .option('ip', {
        alias: 'i',
        describe: `Set server IP address (IPv4 only!).`,
        type: 'string',
        requiresArg: true,
        nargs: 1,
        default: '127.0.0.1:4000'
      })
      .check(argv => {
        if (!isValidIP(argv.ip)) {
          throw new Error(`not a valid IP address: ${argv.ip}`)
        }
        ip = argv.ip
        verbose = argv.verbose
        for (var p = 1; p < argv._.length; p++) {
          const peerIP = argv._[p]
          if (!isValidIP(peerIP)) {
            throw new Error(`not a valid peer IP address: ${peerIP}`)
          }
          if (peerIP !== argv.ip) {
            if (!peers.includes(peerIP)) {
              peers.push(peerIP)
              introduceOurself(peerIP)
            }
          }
        }
        return true
      })
      .strict(true),
  handler: argv => {
    argv._handled = true
    handleRequests(argv)
  }
}

module.exports = server
