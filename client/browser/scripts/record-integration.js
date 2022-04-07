const { createReadStream, createWriteStream } = require('fs')
const { pipeline } = require('stream')
const { promisify } = require('util')
const { createGzip, unzip, deflate } = require('zlib')

const { readdir, readFile } = require('mz/fs')
const shelljs = require('shelljs')

const pipe = promisify(pipeline)

const findRecordingPath = async path => {
  const content = await readdir(path)

  if (content.length === 0) {
    return
  }

  const recording = content.find(element => element === 'recording.har')

  return recording ? `${path}/${recording}` : findRecordingArchivePath(`${path}/${content[0]}`)
}

const compress = async (input, output) => {
  const gzip = createGzip()
  const source = createReadStream(input)
  const destination = createWriteStream(output)
  await pipe(source, gzip, destination)
}

const compressRecordings = async () => {
  const folders = await readdir('./src/integration/__fixtures__')

  for (const folder of folders) {
    const file = await findRecordingPath(`./src/integration/__fixtures__/${folder}`)

    // delete existing recording

    if (file) {
      try {
        // console.log(await readFile(file, 'utf-8'))
        await compress(file, `${file}.gz`)
      } catch (error) {
        console.error('An error occurred:', error)
        process.exitCode = 1
      }
    }
  }
}

const recordSnapshot = grepValue =>
  shelljs.exec(
    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    `POLLYJS_MODE=record SOURCEGRAPH_BASE_URL=https://sourcegraph.com yarn test-integration --grep='${grepValue}'`,
    (error, stdout, stderr) => {
      if (error) {
        console.error(error)
        return
      }
      console.log(`stdout: ${stdout}`)
      console.error(`stderr: ${stderr}`)
    }
  )

;(async () => {
  // 1. Record by --grep args
  const args = process.argv.slice(2)
  for (let index = 0; index < args.length; ++index) {
    if (args[index] === '--grep' && !!args[index + 1]) {
      recordSnapshot(args[index + 1])
      return compressRecordings()
    }
    if (args[index].startsWith('--grep=')) {
      recordSnapshot(args.replace('--grep=', ''))
      return compressRecordings()
    }
  }

  // 2. Record all tests
  const fileNames = await readdir('./src/integration')
  const testFileNames = fileNames.filter(fileName => fileName.endsWith('.test.ts'))
  const testFiles = await Promise.all(
    testFileNames.map(testFileName => readFile(`./src/integration/${testFileName}`, 'utf-8'))
  )

  const testNames = testFiles
    // Ignore template strings for now. If we have lots of tests with parameterized test names, we
    // can use heuristics to still be able to run them.
    .flatMap(testFile => testFile.split('\n').map(line => line.match(/\bit\((["'])(.*)\1/)))
    .filter(Boolean)
    .map(matchArray => matchArray[2])

  for (const testName of testNames) {
    recordSnapshot(testName)
  }

  return compressRecordings()
})().catch(error => {
  console.log(error)
})
