const { readdir, readFile } = require('mz/fs')
const shelljs = require('shelljs')

;(async () => {
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
    shelljs.exec(
      `POLLYJS_MODE=record SOURCEGRAPH_BASE_URL=https://sourcegraph.com yarn test-integration --grep='${testName}'`,
      (error, stdout, stderr) => {
        if (error) {
          console.error(error)
          return
        }
        console.log(`stdout: ${stdout}`)
        console.error(`stderr: ${stderr}`)
      }
    )
  }
})().catch(error => {
  console.log(error)
})
