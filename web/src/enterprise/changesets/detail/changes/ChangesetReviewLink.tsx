import React from 'react'
import { GitPullRequestIcon } from '../../../../util/octicons'
import { GitHubPRLink } from '../../../threads/settings'

interface Props {
    link: GitHubPRLink
    showRepositoryName?: boolean
    showIcon?: boolean
    className?: string
    iconClassName?: string
}

/**
 * A link to a reviewable pull request in a changeset.
 */
export const ChangesetReviewLink: React.FunctionComponent<Props> = ({
    link,
    showRepositoryName,
    showIcon = true,
    className = '',
    iconClassName = 'text-success', // TODO!(sqs): get PR status
}) => (
    <a className={`d-flex align-items-center ${className}`} href={link.url} target="_blank">
        {showIcon && <GitPullRequestIcon className={`icon-inline mr-1 ${iconClassName}`} />}{' '}
        {showRepositoryName &&
            link.repositoryName
                .split('/')
                .slice(1)
                .join('/')}
        #{link.number}
    </a>
)
