import { MdiReactIconProps } from 'mdi-react'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import * as React from 'react'

/**
 * Returns the icon for the repository's code host
 */
export function getRepoIcon(repoName: string): React.ComponentType<MdiReactIconProps> | undefined {
    let RepoIcon: React.ComponentType<MdiReactIconProps> | undefined
    if (repoName.startsWith('github.com/')) {
        RepoIcon = GithubIcon
    }
    if (repoName.startsWith('gitlab.com/')) {
        RepoIcon = GitlabIcon
    }
    if (repoName.startsWith('bitbucket.com/')) {
        RepoIcon = BitbucketIcon
    }

    return RepoIcon
}
