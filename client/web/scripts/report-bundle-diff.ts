const path = require('path')

const shelljs = require('shelljs')

const queryFile = path.join(__dirname, 'report-bundle-jora-query')

const [commitFilename, mergeBaseFilename] = process.argv.slice(-2)
const commitFile = path.join('..', '..', commitFilename)
const mergeBaseFile = path.join('..', '..', mergeBaseFilename)

const statoscope = path.join(__dirname, '..', '..', '..', 'node_modules', '.bin', 'statoscope')

console.log({ queryFile, commitFile, mergeBaseFile, statoscope })

const rawReport = shelljs.exec(`cat "${queryFile}" | ${statoscope} query -i "${commitFile}" -i "${mergeBaseFile}"`)

try {
    const report = JSON.parse(rawReport)

    console.log(report)
    console.log(process.env)
} catch (error) {
    console.log('Error parsing report')
    console.error(error)
    process.exit(1)
}
