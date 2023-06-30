import { isErrorLike } from '@sourcegraph/common'

import { Editor } from '../editor'
import { PreciseContextResult, SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql/client'

export interface GitInfo {
    repo: string
    commitID: string
}

export abstract class GraphContextFetcher {
    // TODO - move into editor interface
    abstract getGitInfo(workspaceRoot: string): Promise<GitInfo>

    constructor(private graphqlClient: SourcegraphGraphQLAPIClient, private editor: Editor) {}

    async getContext(): Promise<PreciseContextResult[]> {
        const editorContext = this.editor.getActiveTextEditor()
        if (!editorContext) {
            return []
        }

        const workspaceRoot = this.editor.getWorkspaceRootPath()
        if (!workspaceRoot) {
            return []
        }

        const { repo: repository, commitID } = await this.getGitInfo(workspaceRoot)
        const activeFile = pathRelativeToRoot(editorContext.filePath, workspaceRoot)

        const response = await this.graphqlClient.getPreciseContext(
            repository,
            commitID,
            activeFile,
            editorContext.content
        )
        if (isErrorLike(response)) {
            return []
        }

        return response
    }
}

function pathRelativeToRoot(path: string, workspaceRoot: string) {
    if (path.startsWith(workspaceRoot)) {
        // +1 for the slash so we produce a relative path
        return path.slice(workspaceRoot.length + 1)
    }

    // Outside of workspace, return absolute file path
    return path
}
