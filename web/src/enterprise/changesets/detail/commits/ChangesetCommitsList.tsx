import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { GitCommitNode } from '../../../../repo/commits/GitCommitNode'
import { useChangesetCommits } from './useChangesetCommits'

interface Props {
    changeset: Pick<GQL.IChangeset, 'id'>

    showCommits?: boolean

    className?: string
}

const LOADING = 'loading' as const

/**
 * A list of commits in a changeset.
 */
export const ChangesetCommitsList: React.FunctionComponent<Props> = ({ changeset, className = '' }) => {
    const commits = useChangesetCommits(changeset)
    return (
        <div className={`changeset-commits-list ${className}`}>
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
