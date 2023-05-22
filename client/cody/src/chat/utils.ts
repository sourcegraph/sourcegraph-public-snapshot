import { spawnSync } from 'child_process'
import path from 'path'

import * as vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { SourcegraphEmbeddingsSearchClient } from '@sourcegraph/cody-shared/src/embeddings/client'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { LocalKeywordContextFetcher } from '../keyword-context/local-keyword-context-fetcher'

import { Config } from './ChatViewProvider'

function filePathContains(container: string, contained: string): boolean {
    let trimmedContained = contained
    if (trimmedContained.endsWith(path.sep)) {
        trimmedContained = trimmedContained.slice(0, -path.sep.length)
    }
    if (trimmedContained.startsWith(path.sep)) {
        trimmedContained = trimmedContained.slice(path.sep.length)
    }
    if (trimmedContained.startsWith('.' + path.sep)) {
        trimmedContained = trimmedContained.slice(1 + path.sep.length)
    }
    return (
        container.includes(path.sep + trimmedContained + path.sep) || // mid-level directory
        container.endsWith(path.sep + trimmedContained) // child
    )
}

export async function filesExist(filePaths: string[]): Promise<{ [filePath: string]: boolean }> {
    const searchPath = `{${filePaths.join(',')}}`
    const realFiles = await vscode.workspace.findFiles(searchPath, null, filePaths.length * 5)
    const ret: { [filePath: string]: boolean } = {}
    for (const filePath of filePaths) {
        let pathExists = false
        for (const realFile of realFiles) {
            if (filePathContains(realFile.fsPath, filePath)) {
                pathExists = true
                break
            }
        }
        ret[filePath] = pathExists
    }
    return ret
}

// Converts a git clone URL to the codebase name that includes the slash-separated code host, owner, and repository name
// This should captures:
// - "github:sourcegraph/sourcegraph" a common SSH host alias
// - "https://github.com/sourcegraph/deploy-sourcegraph-k8s.git"
// - "git@github.com:sourcegraph/sourcegraph.git"
export function convertGitCloneURLToCodebaseName(cloneURL: string): string | null {
    if (!cloneURL) {
        console.error(`Unable to determine the git clone URL for this workspace.\ngit output: ${cloneURL}`)
        return null
    }
    try {
        const uri = new URL(cloneURL.replace('git@', ''))
        // Handle common Git SSH URL format
        const match = cloneURL.match(/git@([^:]+):([\w-]+)\/([\w-]+)(\.git)?/)
        if (cloneURL.startsWith('git@') && match) {
            const host = match[1]
            const owner = match[2]
            const repo = match[3]
            return `${host}/${owner}/${repo}`
        }
        // Handle GitHub URLs
        if (uri.protocol.startsWith('github') || uri.href.startsWith('github')) {
            return `github.com/${uri.pathname.replace('.git', '')}`
        }
        // Handle GitLab URLs
        if (uri.protocol.startsWith('gitlab') || uri.href.startsWith('gitlab')) {
            return `gitlab.com/${uri.pathname.replace('.git', '')}`
        }
        // Handle HTTPS URLs
        if (uri.protocol.startsWith('http') && uri.hostname && uri.pathname) {
            return `${uri.hostname}${uri.pathname.replace('.git', '')}`
        }
        // Generic URL
        if (uri.hostname && uri.pathname) {
            return `${uri.hostname}${uri.pathname.replace('.git', '')}`
        }
        return null
    } catch (error) {
        console.log(`Cody could not extract repo name from clone URL ${cloneURL}:`, error)
        return null
    }
}

export async function getCodebaseContext(
    config: Config,
    rgPath: string,
    editor: Editor
): Promise<CodebaseContext | null> {
    const client = new SourcegraphGraphQLAPIClient(config)
    const workspaceRoot = editor.getWorkspaceRootPath()
    if (!workspaceRoot) {
        return null
    }
    const gitCommand = spawnSync('git', ['remote', 'get-url', 'origin'], { cwd: workspaceRoot })
    const gitOutput = gitCommand.stdout.toString().trim()
    // Get codebase from config or fallback to getting repository name from git clone URL
    const codebase = config.codebase || convertGitCloneURLToCodebaseName(gitOutput)
    if (!codebase) {
        return null
    }
    // Check if repo is embedded in endpoint
    const repoId = await client.getRepoIdIfEmbeddingExists(codebase)
    if (isError(repoId)) {
        const infoMessage = `Cody could not find embeddings for '${codebase}' on your Sourcegraph instance.\n`
        console.info(infoMessage)
        return null
    }

    const embeddingsSearch = repoId && !isError(repoId) ? new SourcegraphEmbeddingsSearchClient(client, repoId) : null
    return new CodebaseContext(config, codebase, embeddingsSearch, new LocalKeywordContextFetcher(rgPath, editor))
}
