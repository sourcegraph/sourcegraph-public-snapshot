import * as React from 'react'

import { mdiBitbucket, mdiGithub, mdiGitlab } from '@mdi/js'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

const iconMap: { [key: string]: { svgPath: string; color?: string } } = {
    'github.com': { svgPath: mdiGithub, color: 'var(--body-color)' },
    'gitlab.com': { svgPath: mdiGitlab, color: '#E24329' },
    'bitbucket.org': { svgPath: mdiBitbucket, color: '#2584FF' },
}
export function codeHostIcon(repoName: string): { hostName: string; svgPath?: string; color?: string } {
    const hostName = repoName.split('/')[0]

    return { hostName, svgPath: iconMap[hostName]?.svgPath, color: iconMap[hostName]?.color }
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
    const { hostName, svgPath, color } = codeHostIcon(repoName)

    if (svgPath) {
        return (
            <Tooltip content={hostName}>
                <Icon aria-label={hostName} className={className} svgPath={svgPath} color={color} />
            </Tooltip>
        )
    }

    return null
}
