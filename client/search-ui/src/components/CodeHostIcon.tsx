import * as React from 'react'

import { mdiBitbucket, mdiGithub, mdiGitlab } from '@mdi/js'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

/**
 * Returns the icon for the repository's code host
 */
export const CodeHostIcon: React.FunctionComponent<
    React.PropsWithChildren<{ repoName: string; className?: string }>
> = ({ repoName, className }) => {
    const iconMap: { [key: string]: string } = {
        'github.com': mdiGithub,
        'gitlab.com': mdiGitlab,
        'bitbucket.org': mdiBitbucket,
    }

    const hostName = repoName.split('/')[0]

    const codehostIconPath: string | undefined = iconMap[hostName]

    if (codehostIconPath) {
        return (
            <Tooltip content={hostName}>
                <Icon aria-label={hostName} className={className} svgPath={codehostIconPath} />
            </Tooltip>
        )
    }

    return null
}
