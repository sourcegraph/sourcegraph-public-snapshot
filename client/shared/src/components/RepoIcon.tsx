import * as React from 'react'

import { MdiReactIconProps } from 'mdi-react'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'

import { Icon } from '@sourcegraph/wildcard'

/**
 * Returns the icon for the repository's code host
 */
export const RepoIcon: React.FunctionComponent<React.PropsWithChildren<{ repoName: string; className?: string }>> = ({
    repoName,
    className,
}) => {
    const iconMap: { [key: string]: React.ComponentType<React.PropsWithChildren<MdiReactIconProps>> } = {
        'github.com': GithubIcon,
        'gitlab.com': GitlabIcon,
        'bitbucket.com': BitbucketIcon,
    }

    const hostName = repoName.split('/')[0]

    const CodehostIcon: React.ComponentType<React.PropsWithChildren<MdiReactIconProps>> | undefined = iconMap[hostName]

    if (CodehostIcon) {
        return (
            <span role="img" aria-label={hostName} title={hostName}>
                <Icon role="img" className={className} as={CodehostIcon} aria-hidden={true} />
            </span>
        )
    }

    return null
}
