const { readdir, readFile } = require('mz/fs')
const shelljs = require('shelljs')

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
      return
    }
    if (args[index].startsWith('--grep=')) {
      recordSnapshot(args.replace('--grep=', ''))
      return
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
})().catch(error => {
  console.log(error)
})
