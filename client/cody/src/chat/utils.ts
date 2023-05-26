import { spawnSync } from 'child_process'
import fs from 'fs'
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
    if (filePaths.length === 0) {
        return {}
    }
    const { debug } = await import('../log')

    const searchPath = `{${filePaths.join(',')}}`
    debug('ChatViewProvider:filesExist', `searchPath: ${searchPath}`)
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

// Walks parent directories from filePath looking for a .git folder.
// If found, returns the folder, otherwise null.
function nearestGitFolder(filePath: string): string | null {
    const components = filePath.split(path.sep)
    components.push()
    do {
        components[components.length - 1] = '.git'
        const pathToCheck = components.join(path.sep)
        const stat = fs.statSync(pathToCheck)
        if (stat.isDirectory()) {
            return pathToCheck
        }
        components.pop()
    } while (components)
    return null
}

// Given a directory, consults git to get the origin URL and converts it to a
// codebase name. Returns undefined if there is none.
function maybeGetCodebaseFromDirectory(dirPath: string): string | undefined {
    const gitCommand = spawnSync('git', ['remote', 'get-url', 'origin'], { cwd: dirPath })
    const gitOutput = gitCommand.stdout.toString().trim()
    return convertGitCloneURLToCodebaseName(gitOutput) || undefined
}

// Given a configuration and editor, returns an object which can produce the
// codebase to attempt to use for embeddings.
export function getCodebaseCandidate(config: Config, editor: Editor): CodebaseCandidate | undefined {
    const editorFilePath = editor.getActiveTextEditor()?.filePath
    const workspaceRoot = editor.getWorkspaceRootPath()
    return new CodebaseCandidate(
        config.codebase,
        (editorFilePath && nearestGitFolder(editorFilePath)) || undefined,
        workspaceRoot || undefined
    )
}

export class CodebaseCandidate {
    constructor(
        private configCodebase: string | undefined,
        private closestGit: string | undefined,
        private workspaceRoot: string | undefined
    ) {}

    // Gets whether this codebase candidate is equivalent to other, given the
    // fallback logic for which codebase to use.
    public isEquivalent(other: CodebaseCandidate | undefined): boolean {
        // This condition must be kept in sync with the codebase getter below so
        // the following condition holds:
        // a.isEquivalent(b) implies a.codebase === b.codebase.
        // This check provides a (relatively) fast path to avoid changing
        // context.
        return (
            !!other &&
            ((this.configCodebase && this.configCodebase === other.configCodebase) ||
                (this.closestGit === other.closestGit && this.workspaceRoot === other.workspaceRoot))
        )
    }

    // Gets the codebase to attempt to use for this candidate.
    public get codebase(): string | undefined {
        return (
            this.configCodebase ||
            (this.closestGit && maybeGetCodebaseFromDirectory(this.closestGit)) ||
            (this.workspaceRoot && maybeGetCodebaseFromDirectory(this.workspaceRoot))
        )
    }
}

// Given a config, ripgrep path and codebase, create the CodebaseContext for it.
// See getCodebaseCandidate(...).codebase to determine the codebase.
export async function getCodebaseContext(
    config: Config,
    rgPath: string,
    editor: Editor,
    codebase: string | undefined
): Promise<CodebaseContext | null> {
    let embeddingsSearch = null
    if (codebase) {
        const client = new SourcegraphGraphQLAPIClient(config)
        // Check if repo is embedded in endpoint
        const repoId = await client.getRepoIdIfEmbeddingExists(codebase)
        if (isError(repoId)) {
            const infoMessage = `Cody could not find embeddings for '${codebase}' on your Sourcegraph instance.\n`
            console.info(infoMessage)
            return null
        }
        embeddingsSearch = repoId && !isError(repoId) ? new SourcegraphEmbeddingsSearchClient(client, repoId) : null
    }
    return new CodebaseContext(config, codebase, embeddingsSearch, new LocalKeywordContextFetcher(rgPath, editor))
}
