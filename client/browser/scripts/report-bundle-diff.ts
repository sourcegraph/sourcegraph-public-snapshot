import path from 'path'

import shelljs from 'shelljs'

const queryFile = path.join(__dirname, 'report-bundle-jora-query')

const [commitFilename, mergeBaseFilename] = process.argv.slice(-2)
const commitFile = path.join('..', '..', commitFilename)
const mergeBaseFile = path.join('..', '..', mergeBaseFilename)

console.log({ queryFile, commitFile, mergeBaseFile })

const rawReport = shelljs.exec(`cat "${queryFile}" | statoscope query -i "${commitFile}" -i "${mergeBaseFile}"`)

const report = JSON.parse(rawReport)

console.log(report)
console.log(process.env)
