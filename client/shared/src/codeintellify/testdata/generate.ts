import * as fs from 'fs'
import * as path from 'path'

import { generateGithubCodeTable } from './github/generate'
import { generateSourcegraphCodeTable } from './sourcegraph/generate'

const generatedDirectory = path.join(__dirname, 'generated')

if (fs.existsSync(generatedDirectory)) {
    fs.rmdirSync(generatedDirectory, { recursive: true })
}
fs.mkdirSync(generatedDirectory, { recursive: true })

const code = fs.readFileSync(path.join(__dirname, 'mux.go.txt'), 'utf-8').split('\n')

fs.writeFileSync(path.join(generatedDirectory, 'github.html'), generateGithubCodeTable(code))
fs.writeFileSync(path.join(generatedDirectory, 'sourcegraph.html'), generateSourcegraphCodeTable(code))
