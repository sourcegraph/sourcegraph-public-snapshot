import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { IssuesList } from '../../list/IssuesList'
import { useIssues } from '../../list/useIssues'
import { RepositoryIssuesAreaContext } from '../RepositoryIssuesArea'

interface Props extends Pick<RepositoryIssuesAreaContext, 'repo'>, ExtensionsControllerNotificationProps {
    newIssueURL: string | null
}

/**
 * Lists a repository's issues.
 */
export const RepositoryIssuesListPage: React.FunctionComponent<Props> = ({ newIssueURL, repo, ...props }) => {
    const issues = useIssues(repo)
    return (
        <div className="repository-issues-list-page">
            {newIssueURL && (
                <Link to={newIssueURL} className="btn btn-primary mb-3">
                    New issue
                </Link>
            )}
            <IssuesList {...props} issues={issues} />
        </div>
    )
}
