import * as fs from 'fs'
import * as path from 'path'

import { generateGithubCodeTable } from './github/generate'
import { generateSourcegraphCodeTable } from './sourcegraph/generate'

const generatedDir = path.join(__dirname, 'generated')

if (fs.existsSync(generatedDir)) {
    fs.rmdirSync(generatedDir, { recursive: true })
}
fs.mkdirSync(generatedDir, { recursive: true })

const code = fs.readFileSync(path.join(__dirname, 'mux.go.txt'), 'utf-8').split('\n')

fs.writeFileSync(path.join(generatedDir, 'github.html'), generateGithubCodeTable(code))
fs.writeFileSync(path.join(generatedDir, 'sourcegraph.html'), generateSourcegraphCodeTable(code))
