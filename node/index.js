#!/usr/bin/env node

const yargs = require('yargs')

try {
  const argv = yargs
    .commandDir('./lib/cmd')
    .option('verbose', {
      alias: 'v',
      description: 'Verbosely list operations performed.',
      type: 'boolean'
    })
    .help()
    .strict(true)
    .fail((msg, err, yargs) => {
      yargs.showHelp('log')
      throw new Error(msg)
    })
    .epilog(`See '${yargs.$0} <command> help' to read about a specific subcommand.`)
    .argv

  if (!argv._handled) {
    yargs.showHelp('log')
    if (argv._[0] !== undefined) {
      throw new Error(`${argv._[0]}: no such command`)
    }
  }
} catch (error) {
  console.error(`error: ${error.message}`)
  process.exit(1)
}
