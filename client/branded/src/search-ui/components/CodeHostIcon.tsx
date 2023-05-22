import * as React from 'react'

import { mdiBitbucket, mdiGithub, mdiGitlab } from '@mdi/js'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

const iconMap: { [key: string]: string } = {
    'github.com': mdiGithub,
    'gitlab.com': mdiGitlab,
    'bitbucket.org': mdiBitbucket,
}
export function codeHostIcon(repoName: string): { hostName: string; svgPath?: string } {
    const hostName = repoName.split('/')[0]

    return { hostName, svgPath: iconMap[hostName] }
}

export function isValidCodeHost(repoName: string): boolean {
    const hostName = repoName.split('/')[0]
    return iconMap[hostName] !== undefined
}

/**
 * Returns the icon for the repository's code host
 */
export const CodeHostIcon: React.FunctionComponent<
    React.PropsWithChildren<{ repoName: string; className?: string }>
> = ({ repoName, className }) => {
    const { hostName, svgPath } = codeHostIcon(repoName)

    if (svgPath) {
        return (
            <Tooltip content={hostName}>
                <Icon aria-label={hostName} className={className} svgPath={svgPath} />
            </Tooltip>
        )
    }

    return null
}
