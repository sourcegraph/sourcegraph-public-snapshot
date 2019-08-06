import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThreadStateIcon } from '../../threadlike/threadState/ThreadStateIcon'

interface Props {
    issue: Pick<GQL.IIssue, '__typename' | 'number' | 'title' | 'state' | 'url'>
}

/**
 * An item in the list of issues.
 */
export const IssueListItem: React.FunctionComponent<Props> = ({ issue }) => (
    <Link to={issue.url} className="d-flex align-items-center text-decoration-none">
        <ThreadStateIcon thread={issue} className="mr-2" />
        <span className="text-muted mr-2">#{issue.number}</span> {issue.title}
    </Link>
)
