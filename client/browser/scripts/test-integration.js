const shelljs = require('shelljs')

const { decompressRecordings, deleteRecordings } = require('./utils')

// eslint-disable-next-line no-void
void (async () => {
  await decompressRecordings()
  shelljs.exec('POLLYJS_MODE=replay yarn run-integration', () => deleteRecordings())
})()
