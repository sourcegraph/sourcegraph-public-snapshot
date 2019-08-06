import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { IssueListItem } from './IssueListItem'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerNotificationProps {
    issues: typeof LOADING | GQL.IIssueConnection | ErrorLike
}

/**
 * Lists issues.
 */
export const IssuesList: React.FunctionComponent<Props> = ({ issues, ...props }) => (
    <div className="issues-list">
        {issues === LOADING ? (
            <LoadingSpinner className="icon-inline mt-3" />
        ) : isErrorLike(issues) ? (
            <div className="alert alert-danger mt-3">{issues.message}</div>
        ) : (
            <div className="card">
                <div className="card-header">
                    <span className="text-muted">
                        {issues.totalCount} {pluralize('issue', issues.totalCount)}
                    </span>
                </div>
                {issues.nodes.length > 0 ? (
                    <ul className="list-group list-group-flush">
                        {issues.nodes.map(issue => (
                            <li key={issue.id} className="list-group-item">
                                <IssueListItem {...props} issue={issue} />
                            </li>
                        ))}
                    </ul>
                ) : (
                    <div className="p-2 text-muted">No issues yet.</div>
                )}
            </div>
        )}
    </div>
)
