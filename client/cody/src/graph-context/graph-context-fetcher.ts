import { spawnSync } from 'child_process'

import { GitInfo, GraphContextFetcher } from '@sourcegraph/cody-shared/src/graph-context'

import { convertGitCloneURLToCodebaseName } from '../chat/utils'

export class VSCodeGraphContextFetcher extends GraphContextFetcher {
    getGitInfo(workspaceRoot: string): Promise<GitInfo> {
        // Get codebase from config or fallback to getting repository name from git clone URL
        const gitCommand = spawnSync('git', ['remote', 'get-url', 'origin'], { cwd: workspaceRoot })
        const gitOutput = gitCommand.stdout.toString().trim()
        const repository = convertGitCloneURLToCodebaseName(gitOutput) || ''
        const codebaseNameSplit = repository.split('/')
        const repoName = codebaseNameSplit.length ? codebaseNameSplit[codebaseNameSplit.length - 1] : ''

        const gitOIDCommand = spawnSync('git', ['rev-parse', 'HEAD'], { cwd: workspaceRoot })
        const commitID = gitOIDCommand.stdout.toString().trim()

        return Promise.resolve({
            repoName,
            commitID,
        })
    }
}
