import { capitalize } from 'lodash'

import type { IconComponent } from '$lib/Icon.svelte'

const iconMap: { [key: string]: IconComponent } = {
    github: ISimpleIconsGithub,
    'github.com': ISimpleIconsGithub,
    gitlab: ISimpleIconsGitlab,
    'gitlab.com': ISimpleIconsGitlab,
    bitbucket: ISimpleIconsBitbucket,
    'bitbucket.org': ISimpleIconsBitbucket,
}

const humanNameMap: { [key: string]: string } = {
    github: 'GitHub',
    'github.com': 'GitHub',
    gitlab: 'GitLab',
    'gitlab.com': 'GitLab',
    bitbucket: 'Bitbucket',
    'bitbucket.org': 'Bitbucket',
}

/**
 * Returns the SVG icon component for the given code host. Accepts the code host's name
 * (e.g. 'github') or hostname  (e.g. 'github.com'). Defaults to a generic source
 * repository icon.
 *
 * @param codeHost The code host name or hostname.
 * @returns The SVG icon component for the code host.
 */
export function getIconForCodeHost(codeHost: string): IconComponent {
    return iconMap[codeHost.toLowerCase()] ?? ILucideFolderGit2
}

/**
 * Returns the human-readable name for the given code host. Accepts the code host's name
 * (e.g. 'github') or hostname  (e.g. 'github.com'). Defaults to capitalizing
 * the name.
 *
 * @param codeHost The code host name or hostname.
 * @returns The human-readable name for the code host.
 */
export function getHumanNameForCodeHost(codeHost: string): string {
    return humanNameMap[codeHost.toLowerCase()] ?? capitalize(codeHost)
}
