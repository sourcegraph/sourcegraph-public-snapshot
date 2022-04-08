const shelljs = require('shelljs')

const { decompressRecordings, deleteRecordings } = require('./utils')

;(async () => {
  await decompressRecordings()
  shelljs.exec('POLLYJS_MODE=replay yarn run-integration', () => deleteRecordings())
})().catch(error => {
  console.log(error)
})
