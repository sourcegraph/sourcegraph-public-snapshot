import { spawnSync } from 'child_process'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { SourcegraphEmbeddingsSearchClient } from '@sourcegraph/cody-shared/src/embeddings/client'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { LocalKeywordContextFetcher } from '../keyword-context/local-keyword-context-fetcher'

import { Config } from './ChatViewProvider'
import { AuthStatus, defaultAuthStatus, isLocalApp, unauthenticatedStatus } from './protocol'

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

let client: SourcegraphGraphQLAPIClient
let configWithToken: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>

/**
 * Gets the authentication status for a user.
 *
 * @returns The user's authentication status.
 */
export async function getAuthStatus(
    config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
): Promise<AuthStatus> {
    // Cache the config and the GraphQL client
    if (config !== configWithToken) {
        configWithToken = config
        client = new SourcegraphGraphQLAPIClient(config)
    }
    // Early return if no access token
    if (!config.accessToken) {
        return { ...unauthenticatedStatus }
    }
    // Version is for frontend to check if Cody is not enabled due to unsupported version when siteHasCodyEnabled is false
    const { enabled, version } = await client.isCodyEnabled()
    // Use early returns to avoid nested if-else
    const isDotComOrApp = client.isDotCom() || isLocalApp(config.serverEndpoint)
    if (isDotComOrApp) {
        const userInfo = await client.getCurrentUserIdAndVerifiedEmail()
        if (!userInfo || isError(userInfo)) {
            return { ...unauthenticatedStatus }
        }
        if (userInfo.id.length > 0) {
            return newAuthStatus(isDotComOrApp, userInfo.id, userInfo.hasVerifiedEmail, true, version)
        }
        return { ...unauthenticatedStatus }
    }
    const userId = await client.getCurrentUserId()
    if (!userId || isError(userId)) {
        return { ...unauthenticatedStatus }
    }

    return newAuthStatus(isDotComOrApp, userId, false, enabled, version)
}

/**
 * Checks a user's authentication status.
 *
 * @param isDotComOrApp Whether the user is on an insider build instance or enterprise instance.
 * @param userId The user's ID.
 * @param isEmailVerified Whether the user has verified their email. Default to true for non-enterprise instances.
 * @param isCodyEnabled Whether Cody is enabled on the Sourcegraph instance. Default to true for non-enterprise instances.
 * @param version The Sourcegraph instance version.
 * @returns The user's authentication status. It's for frontend to display when instance is on unsupported version if siteHasCodyEnabled is false
 */
export function newAuthStatus(
    isDotComOrApp: boolean,
    userId: string,
    isEmailVerified: boolean,
    isCodyEnabled: boolean,
    version: string
): AuthStatus {
    if (!userId) {
        return { ...unauthenticatedStatus }
    }
    const newAuthStatus = { ...defaultAuthStatus }
    // Set values and return early
    newAuthStatus.authenticated = !!userId
    newAuthStatus.showInvalidAccessTokenError = !userId
    newAuthStatus.requiresVerifiedEmail = isDotComOrApp
    newAuthStatus.hasVerifiedEmail = isDotComOrApp && isEmailVerified
    newAuthStatus.siteHasCodyEnabled = isCodyEnabled
    newAuthStatus.siteVersion = version
    return newAuthStatus
}
