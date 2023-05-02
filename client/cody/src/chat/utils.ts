import { spawnSync } from 'child_process'

import * as vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { SourcegraphEmbeddingsSearchClient } from '@sourcegraph/cody-shared/src/embeddings/client'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { LocalKeywordContextFetcher } from '../keyword-context/local-keyword-context-fetcher'

import { Config } from './protocol'

export async function isValidLogin(
    config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
): Promise<boolean> {
    const client = new SourcegraphGraphQLAPIClient(config)
    const userId = await client.getCurrentUserId()
    return !isError(userId)
}

export function trimPrefix(text: string, prefix: string): string {
    if (text.startsWith(prefix)) {
        return text.slice(prefix.length)
    }
    return text
}

export function trimSuffix(text: string, suffix: string): string {
    if (text.endsWith(suffix)) {
        return text.slice(0, -suffix.length)
    }
    return text
}

export async function fileExists(filePath: string): Promise<boolean> {
    const patterns = [filePath, '**/' + trimSuffix(trimPrefix(filePath, '/'), '/') + '/**']
    if (!filePath.endsWith('/')) {
        patterns.push('**/' + trimPrefix(filePath, '/') + '*')
    }
    for (const pattern of patterns) {
        const files = await vscode.workspace.findFiles(pattern, null, 1)
        if (files.length > 0) {
            return true
        }
    }
    return false
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
        console.error(`Cody could not extract repo name from clone URL ${cloneURL}:`, error)
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
