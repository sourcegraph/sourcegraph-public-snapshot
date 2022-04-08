const shelljs = require('shelljs')

const { decompressRecordings, deleteRecordings } = require('./utils')

;(async () => {
  await decompressRecordings()
  await new Promise(resolve => shelljs.exec('yarn run-integration', () => resolve()))
  await deleteRecordings()
})().catch(error => {
  console.log(error)
})
