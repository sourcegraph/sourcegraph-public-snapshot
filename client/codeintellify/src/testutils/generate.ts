import * as fs from 'fs'
import * as path from 'path'

import { generateGithubCodeTable } from './github/generate'
import { generateSourcegraphCodeTable } from './sourcegraph/generate'

const code = fs.readFileSync(path.join(__dirname, 'mux.go.txt'), 'utf-8').split('\n')

export const GITHUB_CODE_TABLE = generateGithubCodeTable(code)
export const SOURCEGRAPH_CODE_TABLE = generateSourcegraphCodeTable(code)
