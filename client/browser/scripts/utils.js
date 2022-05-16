const { pipeline } = require('stream')
const { promisify } = require('util')
const { createGzip, unzip } = require('zlib')

const { readdir, writeFile, readFile, unlink, createReadStream, createWriteStream } = require('mz/fs')

const fixturesPath = './src/integration/__fixtures__'
const recordingFileName = 'recording.har'

// eslint-disable-next-line @typescript-eslint/restrict-template-expressions
const buildCompressedFilePath = filePath => `${filePath}.gz`
const buildDecompressedFilePath = compressedFilePath => compressedFilePath.replace(/\.gz$/, '')

const findRecordingPath = async (path, isCompressed) => {
  const content = await readdir(path, { withFileTypes: true })

  if (content.length === 0) {
    return
  }

  if (content[0].isDirectory()) {
    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    return findRecordingPath(`${path}/${content[0].name}`, isCompressed)
  }

  const recording = content.find(
    element =>
      element.isFile() &&
      element.name === (isCompressed ? buildCompressedFilePath(recordingFileName) : recordingFileName)
  )

  if (recording) {
    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    return `${path}/${recording.name}`
  }
}

const pipe = promisify(pipeline)
const compress = async (input, output) => {
  const gzip = createGzip()
  const source = createReadStream(input)
  const destination = createWriteStream(output)
  await pipe(source, gzip, destination)
}

const compressRecordings = async () => {
  const folders = await readdir(fixturesPath)

  await Promise.all(
    folders.map(async folder => {
      const filePath = await findRecordingPath(`${fixturesPath}/${folder}`, false)

      if (filePath) {
        try {
          await compress(filePath, buildCompressedFilePath(filePath))
          await unlink(filePath) // delete original recording
        } catch (error) {
          console.error('An error occurred:', error)
          process.exitCode = 1
        }
      }
    })
  )
}

const unzipAsPromise = promisify(unzip)
const decompressRecordings = async () => {
  const folders = await readdir(fixturesPath)

  await Promise.all(
    folders.map(async folder => {
      const filePath = await findRecordingPath(`${fixturesPath}/${folder}`, true)

      if (filePath) {
        try {
          const content = await readFile(filePath, 'base64')
          const result = await unzipAsPromise(Buffer.from(content, 'base64'))
          await writeFile(buildDecompressedFilePath(filePath), result)
        } catch (error) {
          console.error('An error occurred:', error)
          process.exitCode = 1
        }
      }
    })
  )
}

const deleteRecordings = async () => {
  const folders = await readdir(fixturesPath)

  await Promise.all(
    folders.map(async folder => {
      const filePath = await findRecordingPath(`${fixturesPath}/${folder}`, false)

      if (filePath) {
        try {
          await unlink(filePath) // delete original recording
        } catch (error) {
          console.error('An error occurred:', error)
          process.exitCode = 1
        }
      }
    })
  )
}

module.exports = { compressRecordings, decompressRecordings, deleteRecordings }
