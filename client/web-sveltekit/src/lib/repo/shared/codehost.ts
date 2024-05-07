import { mdiBitbucket, mdiGithub, mdiGitlab, mdiSourceRepository } from '@mdi/js'
import { capitalize } from 'lodash'

const iconMap: { [key: string]: string } = {
    github: mdiGithub,
    'github.com': mdiGithub,
    gitlab: mdiGitlab,
    'gitlab.com': mdiGitlab,
    bitbucket: mdiBitbucket,
    'bitbucket.org': mdiBitbucket,
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
 * Returns the SVG icon path for the given code host. Accepts the code host's name
 * (e.g. 'github') or hostname  (e.g. 'github.com'). Defaults to a generic source
 * repository icon.
 *
 * @param codeHost The code host name or hostname.
 * @returns The SVG icon path for the code host.
 */
export function getIconPathForCodeHost(codeHost: string): string {
    return iconMap[codeHost.toLowerCase()] ?? mdiSourceRepository
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
