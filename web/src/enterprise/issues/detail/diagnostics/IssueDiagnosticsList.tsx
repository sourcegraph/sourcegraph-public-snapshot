import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'

interface Props {
    issue: Pick<GQL.IIssue, 'id'>

    className?: string
}

const LOADING = 'loading' as const

/**
 * A list of diagnostics in an issue.
 */
export const IssueDiagnosticsList: React.FunctionComponent<Props> = ({ issue, className = '' }) => {
    const diagnostics = useIssueDiagnostics(issue)
    return (
        <div className={`issue-diagnostics-list ${className}`}>
            <ul className="list-group mb-4">
                {diagnostics === LOADING ? (
                    <LoadingSpinner className="icon-inline mt-3" />
                ) : isErrorLike(diagnostics) ? (
                    <div className="alert alert-danger mt-3">{diagnostics.message}</div>
                ) : (
                    diagnostics.map((diagnostic, i) => (
                        <li key={i} className="list-group-item p-0">
                            TODO!(sqs) show diagnostic
                        </li>
                    ))
                )}
            </ul>
        </div>
    )
}
