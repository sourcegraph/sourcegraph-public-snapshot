import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { GitCommitNode } from '../../../../repo/commits/GitCommitNode'
import { useThreadCommits } from './useThreadCommits'

interface Props {
    thread: Pick<GQL.IThread, 'id'>

    showCommits?: boolean

    className?: string
}

const LOADING = 'loading' as const

/**
 * A list of commits in a thread.
 */
export const ThreadCommitsList: React.FunctionComponent<Props> = ({ thread, className = '' }) => {
    const commits = useThreadCommits(thread)
    return (
        <div className={`thread-commits-list ${className}`}>
            <ul className="list-group mb-4">
                {commits === LOADING ? (
                    <LoadingSpinner className="icon-inline mt-3" />
                ) : isErrorLike(commits) ? (
                    <div className="alert alert-danger mt-3">{commits.message}</div>
                ) : (
                    commits.map((commit, i) => (
                        <li key={i} className="list-group-item p-0">
                            <GitCommitNode node={commit} compact={true} />
                        </li>
                    ))
                )}
            </ul>
        </div>
    )
}
