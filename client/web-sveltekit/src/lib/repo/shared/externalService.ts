import { ExternalServiceKind } from '$lib/graphql-types'
import { type IconComponent } from '$lib/Icon.svelte'

const iconMap: Partial<Record<keyof typeof ExternalServiceKind, IconComponent>> = {
    GITHUB: ISimpleIconsGithub,
    GITLAB: ISimpleIconsGitlab,
    BITBUCKETCLOUD: ISimpleIconsBitbucket,
}

/**
 * Returns the SVG icon component for the given code host. Defaults to a
 * generic source repository icon.
 *
 * @param codeHost The code host kind.
 * @returns The SVG icon component for the code host.
 */
export function getIconForExternalService(kind: ExternalServiceKind | undefined): IconComponent {
    return kind ? iconMap[kind] : ILucideFolderGit2
}

const humanNameMap: Record<ExternalServiceKind, string> = {
    AWSCODECOMMIT: 'AWS CodeCommit',
    AZUREDEVOPS: 'Azure DevOps',
    BITBUCKETCLOUD: 'Bitbucket Cloud',
    BITBUCKETSERVER: 'Bitbucket Server',
    GERRIT: 'Gerrit',
    GITHUB: 'GitHub',
    GITLAB: 'GitLab',
    GITOLITE: 'Gitolite',
    GOMODULES: 'Go modules',
    JVMPACKAGES: 'JVM packages',
    NPMPACKAGES: 'NPM packages',
    PAGURE: 'Pagure',
    PERFORCE: 'Perforce',
    PHABRICATOR: 'Phabricator',
    PYTHONPACKAGES: 'Python packages',
    RUBYPACKAGES: 'Ruby packages',
    RUSTPACKAGES: 'Rust packages',
    OTHER: 'Unknown',
}

/**
 * Returns the human-readable name for the given code host. Accepts the code host's name
 * (e.g. 'github') or hostname  (e.g. 'github.com'). Defaults to capitalizing
 * the name.
 *
 * @param codeHost The code host kind.
 * @returns The human-readable name for the code host.
 */
export function getHumanNameForExternalService(codeHost: ExternalServiceKind): string {
    return humanNameMap[codeHost]
}

const inferredKindMap: Record<string, ExternalServiceKind> = {
    'github.com': ExternalServiceKind.GITHUB,
    'gitlab.org': ExternalServiceKind.GITLAB,
    'bitbucket.com': ExternalServiceKind.BITBUCKETCLOUD,
}

/**
 * @deprecated Prefer getting the code host kind from the GraphQL API Repository.externalLinks because
 * it is more accurate and will return the correct CodeHostKind for a non-public deployment.
 *
 * Attempts to infer the code host kind from a repo name.
 *
 * @param repoName The name of the repo
 * @returns The inferred service kind, or undefined if unsuccessful
 */
export function inferExternalServiceKind(repoName: string): ExternalServiceKind | undefined {
    const name = repoName.split('/', 2)[0]
    return inferredKindMap[name]
}
