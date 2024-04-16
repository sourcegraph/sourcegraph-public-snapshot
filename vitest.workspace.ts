import { readFileSync } from 'fs'
import path from 'path'

import { load } from 'js-yaml'

interface PnpmWorkspaceFile {
    packages: string[]
}

function fromPnpmWorkspaceFile(filePath: string): string[] {
    return (load(readFileSync(filePath, 'utf8')) as PnpmWorkspaceFile).packages.map(p => `${p}/vitest.config.ts`, {
        cwd: __dirname,
    })
}

export default fromPnpmWorkspaceFile(path.join(__dirname, 'pnpm-workspace.yaml'))
