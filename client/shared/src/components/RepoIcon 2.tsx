import { MdiReactIconProps } from 'mdi-react'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import * as React from 'react'

/**
 * Returns the icon for the repository's code host
 */
export const RepoIcon: React.FunctionComponent<{ repoName: string; className?: string }> = ({
    repoName,
    className,
}) => {
    const iconMap: { [key: string]: React.ComponentType<MdiReactIconProps> } = {
        'github.com': GithubIcon,
        'gitlab.com': GitlabIcon,
        'bitbucket.com': BitbucketIcon,
    }

    const hostName = repoName.split('/')[0]

    const Icon: React.ComponentType<MdiReactIconProps> | undefined = iconMap[hostName]

    if (Icon) {
        return (
            <span aria-label={hostName} title={hostName}>
                <Icon className={className} />
            </span>
        )
    }

    return null
}
