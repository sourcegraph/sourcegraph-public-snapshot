const { promisify } = require('util')
const { unzip } = require('zlib')

const { readdir, readFile, writeFile } = require('mz/fs')
const shelljs = require('shelljs')

const do_unzip = promisify(unzip)

const findRecordingArchivePath = async path => {
  const content = await readdir(path)

  if (content.length === 0) {
    return
  }

  const archive = content.find(element => element.endsWith('gz'))

  return archive ? `${path}/${archive}` : findRecordingArchivePath(`${path}/${content[0]}`)
}

// eslint-disable-next-line no-void
void (async () => {
  const folders = await readdir('./src/integration/__fixtures__')

  for (const folder of folders) {
    const file = await findRecordingArchivePath(`./src/integration/__fixtures__/${folder}`)

    if (file) {
      try {
        const content = await readFile(file, 'base64')
        const result = await do_unzip(Buffer.from(content, 'base64'))
        await writeFile(file.replace(/\.gz$/, ''), result)
      } catch (error) {
        console.error('An error occurred:', error)
        process.exitCode = 1
      }
    }
  }

  shelljs.exec(
    'TS_NODE_PROJECT=src/integration/tsconfig.json SOURCEGRAPH_BASE_URL=https://sourcegraph.com mocha --parallel=$CI --retries=2 ./src/integration/**/*.test.ts'
  )
})()
