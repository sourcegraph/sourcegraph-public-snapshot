import { isErrorLike } from '@sourcegraph/common'

import { Editor } from '../editor'
import { PreciseContextResult, SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql/client'

export interface GitInfo {
    repoName: string
    commitID: string
}

export abstract class GraphContextFetcher {
    constructor(private graphqlClient: SourcegraphGraphQLAPIClient, private editor: Editor) {}

    // NOTE(auguste): maybe we should refactor this into Editor
    abstract getGitInfo(workspaceRoot: string): Promise<GitInfo>

    async getContext(): Promise<PreciseContextResult[]> {
        const active = this.editor.getActiveTextEditor()

        if (!active) {
            return []
        }

        let preciseContext: PreciseContextResult[] = []
        const workspaceRoot = this.editor.getWorkspaceRootPath()
        if (workspaceRoot) {
            const { repoName, commitID } = await this.getGitInfo(workspaceRoot)

            const activeFile = trimPath(active.filePath, repoName)
            const response = await this.graphqlClient.getPreciseContext(repoName, commitID, activeFile, active.content)
            if (!isErrorLike(response)) {
                preciseContext = response
            }
        }

        return preciseContext
    }
}

function trimPath(path: string, folderName: string) {
    const folderIndex = path.indexOf(folderName)

    if (folderIndex === -1) {
        return path
    }

    // Add folderName.length for length of folder name and +1 for the slash
    return path.slice(Math.max(0, folderIndex + folderName.length + 1))
}
