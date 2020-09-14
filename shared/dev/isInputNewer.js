const { exists, stat } = require('mz/fs')
const glob = require('glob')

// Returns true if ay of the files matched by inputGlobs is newer than the outfile.
async function isInputNewer(inputGlobs, outfile) {
  if (!(await exists(outfile))) {
    return true
  }

  const outfileModifiedTime = (await stat(outfile)).mtimeMs
  const infileModifiedTimes = await Promise.all(
    inputGlobs
      .map(inputGlob => glob.sync(inputGlob))
      .reduce((a, b) => a.concat(b))
      .map(async file => (await stat(file)).mtimeMs)
  )
  const maxInTime = infileModifiedTimes.reduce((a, b) => (a > b ? a : b), 0)
  if (maxInTime > outfileModifiedTime) {
    return true
  }
  return false
}

module.exports = isInputNewer
