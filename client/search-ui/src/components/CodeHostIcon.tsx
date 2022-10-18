import * as React from 'react'

import { mdiBitbucket, mdiGithub, mdiGitlab } from '@mdi/js'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

export function codeHostIcon(repoName: string): { hostName: string; svgPath?: string } {
    const hostName = repoName.split('/')[0]
    const iconMap: { [key: string]: string } = {
        'github.com': mdiGithub,
        'gitlab.com': mdiGitlab,
        'bitbucket.org': mdiBitbucket,
    }
    return { hostName, svgPath: iconMap[hostName] }
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
