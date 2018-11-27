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

let verbose = false

const explain = msg => {
  if (verbose) {
    console.log(`${Date.now()}: ${msg}`)
  }
}

const doGet = (ip, key) => {
  const grpcClient = new KVProto.Client(ip, grpc.credentials.createInsecure())
  explain(`sending GET request to ${ip} for key '${key}'...`)
  grpcClient.get({ key }, function (err, response) {
    if (err) {
      console.error(`error: ${err.message}`)
    } else {
      if (response.defined) {
        console.log(`'${key}'='${response.value}'`)
      } else {
        console.log(`'${key}': undefined`)
      }
    }
  })
}

const doSet = (ip, key, value) => {
  const grpcClient = new KVProto.Client(ip, grpc.credentials.createInsecure())
  explain(`sending SET request to ${ip} for key '${key}'...`)
  grpcClient.set({ key, value, broadcast: true }, function (err, response) {
    if (err) {
      console.error(`error: ${err.message}`)
    }
  })
}

const doList = ip => {
  const grpcClient = new KVProto.Client(ip, grpc.credentials.createInsecure())
  explain(`sending LIST request to ${ip}...`)
  grpcClient.list({}, function (err, response) {
    if (err) {
      console.error(`error: ${err.message}`)
    } else {
      console.log(`Key-value pairs defined on ${ip}:`)
      for (const key in response.store) {
        console.log(` - '${key}'='${response.store[key]}'`)
      }
      console.log('-- end of key-value dump --')
    }
  })
}

const handleGet = (ip, key) => {
  const re = /^[a-zA-Z0-9_]+$/
  if (!re.test(key)) {
    throw new Error(`invalid --get: expected '--get KEY'; got '--get ${key}'`)
  }
  doGet(ip, key)
}

const handleSet = (ip, kv) => {
  const re = /^([a-zA-Z0-9_]+)=([a-zA-Z0-9_]+)$/
  const match = re.exec(kv)
  if (match === null) {
    throw new Error(`invalid --set: expected '--set KEY=VALUE'; got '--set ${kv}'`)
  }
  doSet(ip, match[1], match[2])
}

const client = {
  desc: `Send request(s) to a key-value store server.`,
  builder: yargs =>
    yargs
      .option('ip', {
        alias: 'i',
        describe: `Set server IP address (IPv4 only!).`,
        type: 'string',
        default: '127.0.0.1:4000',
        nargs: 1
      })
      .option('get', {
        alias: 'g',
        describe: `Get value associated with key.`,
        type: 'string',
        nargs: 1
      })
      .option('set', {
        alias: 's',
        describe: `Set a key-value pair.`,
        type: 'string',
        nargs: 1
      })
      .option('list', {
        alias: 'l',
        describe: `Get key-value pairs defined on server.`,
        type: 'bool',
        default: false,
        nargs: 0
      })
      // .demandOption(['get', 'set', 'list'])
      .check(argv => {
        if (Array.isArray(argv.ip)) {
          argv.ip = argv.ip[argv.ip.length - 1]
          console.log(`connecting to server ${argv.ip}...`)
        }
        if (!isValidIP(argv.ip)) {
          throw new Error(`not a valid server IP address: ${argv.ip}`)
        }
        verbose = argv.verbose
        return true
      })
      .strict(true),
  handler: argv => {
    argv._handled = true
    let cmdExecuted = false
    if (typeof argv.get !== 'undefined') {
      cmdExecuted = true
      if (argv.get instanceof Array) {
        argv.get.forEach(key => handleGet(argv.ip, key))
      } else {
        handleGet(argv.ip, argv.get)
      }
    }
    if (typeof argv.set !== 'undefined') {
      cmdExecuted = true
      if (argv.set instanceof Array) {
        argv.set.forEach(kv => handleSet(argv.ip, kv))
      } else {
        handleSet(argv.ip, argv.set)
      }
    }
    if (argv.list) {
      cmdExecuted = true
      doList(argv.ip)
    }
    if (!cmdExecuted) {
      console.log('nothing to do; please specify an operation')
    }
  }
}

module.exports = client
