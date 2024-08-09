import { ExternalServiceKind } from '$lib/graphql-types'
import type { IconComponent } from '$lib/Icon.svelte'

const icons: Partial<Record<ExternalServiceKind, IconComponent>> = {
    GITHUB: ISimpleIconsGithub,
    GITLAB: ISimpleIconsGitlab,
    BITBUCKETSERVER: ISimpleIconsBitbucket,
    BITBUCKETCLOUD: ISimpleIconsBitbucket,
}

/**
 * Returns the SVG icon component for the given code host.
 */
export function getIconForServiceKind(kind: ExternalServiceKind): IconComponent {
    return icons[kind] ?? ILucideFolderGit2
}

const humanNames: Record<ExternalServiceKind, string> = {
    AWSCODECOMMIT: 'AWS CodeCommit',
    AZUREDEVOPS: 'Azure DevOps',
    BITBUCKETCLOUD: 'Bitbucket',
    BITBUCKETSERVER: 'Bitbucket',
    GERRIT: 'Gerrit',
    GITHUB: 'GitHub',
    GITLAB: 'GitLab',
    GITOLITE: 'Gitolite',
    GOMODULES: 'Go Module',
    JVMPACKAGES: 'JVM Packages',
    NPMPACKAGES: 'NPM Packages',
    PAGURE: 'Pagure',
    PERFORCE: 'Perforce',
    PHABRICATOR: 'Phabricator',
    PYTHONPACKAGES: 'Python Packages',
    RUBYPACKAGES: 'Ruby Packages',
    RUSTPACKAGES: 'Rust Packages',
    OTHER: 'Unknown',
}

/**
 * Returns the human-readable name for the given external service kind.
 */
export function getExternalServiceHumanName(kind: ExternalServiceKind): string {
    return humanNames[kind]
}

const knownPrefixes: Partial<Record<string, ExternalServiceKind>> = {
    'github.com': ExternalServiceKind.GITHUB,
    'gitlab.com': ExternalServiceKind.GITLAB,
    'bitbucket.com': ExternalServiceKind.BITBUCKETCLOUD,
}

const splitRepoNameRegex = /^(?<codehost>[^\.\/]+\.[^\.\/]+)\/(?<repo>.*)$/

/**
 * Attempts to infer the code host kind from the repo name and split the repo
 * name into a code host part and a display name part.
 *
 * If provided, the external service kind will be used instead of inferring the
 * kind from the repo name. Over time, we should attempt to provide a kind in
 * every place we display a repo name since guessing will not work for any
 * non-well-known repos.
 *
 * NOTE: there is no guarantee that the first element of a repo name is the
 * code host. Admins can name their repos whatever they want, so it's possible
 * we will incorrectly split if the chosen name matches the regex pattern.
 * We could theoretically use the external URLs to check this, but we do not
 * always have that information handy so we don't do that now.
 */
export function inferSplitCodeHost(
    repoName: string,
    kind: ExternalServiceKind | undefined
): { kind: ExternalServiceKind; codeHost: string; displayName: string } {
    const matched = splitRepoNameRegex.exec(repoName)
    const codeHost = matched?.groups?.codehost ?? ''
    return {
        kind: kind ?? knownPrefixes[codeHost] ?? ExternalServiceKind.OTHER,
        codeHost,
        displayName: matched?.groups?.repo ?? repoName,
    }
}
