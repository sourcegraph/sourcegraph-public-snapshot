import type { IconComponent } from '$lib/Icon.svelte'

const codeHostKinds = ['github', 'gitlab', 'bitbucket'] as const
export type CodeHostKind = typeof codeHostKinds[number]

const iconMap: Record<CodeHostKind, IconComponent> = {
    github: ISimpleIconsGithub,
    gitlab: ISimpleIconsGitlab,
    bitbucket: ISimpleIconsBitbucket,
}

/**
 * Returns the SVG icon component for the given code host. Defaults to a
 * generic source repository icon.
 *
 * @param codeHost The code host kind.
 * @returns The SVG icon component for the code host.
 */
export function getIconForCodeHost(codeHost: CodeHostKind | undefined): IconComponent {
    return codeHost ? iconMap[codeHost] : ILucideFolderGit2
}

const humanNameMap: Record<CodeHostKind, string> = {
    github: 'GitHub',
    gitlab: 'GitLab',
    bitbucket: 'Bitbucket',
}

/**
 * Returns the human-readable name for the given code host. Accepts the code host's name
 * (e.g. 'github') or hostname  (e.g. 'github.com'). Defaults to capitalizing
 * the name.
 *
 * @param codeHost The code host kind.
 * @returns The human-readable name for the code host.
 */
export function getHumanNameForCodeHost(codeHost: CodeHostKind): string {
    return humanNameMap[codeHost]
}

const inferredKindMap: Record<string, CodeHostKind> = {
    'github.com': 'github',
    'gitlab.org': 'gitlab',
    'bitbucket.com': 'bitbucket',
}

/**
 * @deprecated Prefer getting the code host kind from the GraphQL API because
 * it is more accurate and will return the correct CodeHostKind for a non-public deployment.
 *
 * Attempts to infer the code host kind from a repo name.
 *
 * @param repoName The name of the repo
 * @returns An object containing an inferred name and kind for the repo's code host
 */
export function inferCodeHost(repoName: string): { name: string; kind: CodeHostKind | undefined } {
    const name = repoName.split(repoName, 2)[0]
    return { name, kind: inferredKindMap[name] }
}
