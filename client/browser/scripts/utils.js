const { pipeline } = require('stream')
const { promisify } = require('util')
const { createGzip, unzip } = require('zlib')

const { readdir, writeFile, readFile, unlink, createReadStream, createWriteStream } = require('mz/fs')

const fixturesPath = './src/integration/__fixtures__'
const recordingFileName = 'recording.har'

const buildCompressedFilePath = filePath => `${filePath}.gz`
const buildDecompressedFilePath = compressedFilePath => compressedFilePath.replace(/\.gz$/, '')

const findRecordingPath = async (path, isCompressed) => {
  const content = await readdir(path)

  if (content.length === 0) {
    return
  }

  const recording = content.find(
    element => element === (isCompressed ? buildCompressedFilePath(recordingFileName) : recordingFileName)
  )

  return recording ? `${path}/${recording}` : findRecordingPath(`${path}/${content[0]}`, isCompressed)
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

  for (const folder of folders) {
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
  }
}

const unzipAsPromise = promisify(unzip)
const decompressRecordings = async () => {
  const folders = await readdir(fixturesPath)

  for (const folder of folders) {
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
  }
}

const deleteRecordings = async () => {
  const folders = await readdir(fixturesPath)

  for (const folder of folders) {
    const file = await findRecordingPath(`${fixturesPath}/${folder}`, false)

    if (file) {
      try {
        await unlink(file) // delete original recording
      } catch (error) {
        console.error('An error occurred:', error)
        process.exitCode = 1
      }
    }
  }
}

module.exports = { compressRecordings, decompressRecordings, deleteRecordings }
