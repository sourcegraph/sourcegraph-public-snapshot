import { AuthStatus, defaultAuthStatus, unauthenticatedStatus } from './protocol'

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
    endpoint: string,
    isDotComOrApp: boolean,
    user: boolean,
    isEmailVerified: boolean,
    isCodyEnabled: boolean,
    version: string,
    configOverwrites?: AuthStatus['configOverwrites']
): AuthStatus {
    if (!user) {
        return { ...unauthenticatedStatus, endpoint }
    }
    const authStatus: AuthStatus = { ...defaultAuthStatus, endpoint }
    // Set values and return early
    authStatus.authenticated = user
    authStatus.showInvalidAccessTokenError = !user
    authStatus.requiresVerifiedEmail = isDotComOrApp
    authStatus.hasVerifiedEmail = isDotComOrApp && isEmailVerified
    authStatus.siteHasCodyEnabled = isCodyEnabled
    authStatus.siteVersion = version
    if (configOverwrites) {
        authStatus.configOverwrites = configOverwrites
    }
    const isLoggedIn = authStatus.siteHasCodyEnabled && authStatus.authenticated
    const isAllowed = authStatus.requiresVerifiedEmail ? authStatus.hasVerifiedEmail : true
    authStatus.isLoggedIn = isLoggedIn && isAllowed
    return authStatus
}
