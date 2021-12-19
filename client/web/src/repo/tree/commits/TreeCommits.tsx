import classNames from 'classnames'
import { formatISO, subYears } from 'date-fns'
import React, { useCallback, useState } from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'

import { useConnection } from '../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    SummaryContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ShowMoreButton,
} from '../../../components/FilteredConnection/ui'
import { GitCommitFields, TreeCommitsResult, TreeCommitsVariables } from '../../../graphql-operations'
import { GitCommitNode } from '../../commits/GitCommitNode'

import { TREE_COMMITS } from './gql'
import styles from './TreeCommits.module.scss'

interface Props {
    repoID: Scalars['ID']
    commitID: string
    filePath: string
    className?: string
}

const DEFAULT_FIRST = 7

/**
 * A list of commits in a tree (or in the entire repository for the root tree).
 */
export const TreeCommits: React.FunctionComponent<Props> = ({ repoID, commitID, filePath, className }) => {
    const [showOlderCommits, setShowOlderCommits] = useState(false)

    const onShowOlderCommitsClick = useCallback(
        (event: React.MouseEvent): void => {
            event.preventDefault()
            setShowOlderCommits(true)
        },
        [setShowOlderCommits]
    )

    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        TreeCommitsResult,
        TreeCommitsVariables,
        GitCommitFields
    >({
        query: TREE_COMMITS,
        variables: {
            first: DEFAULT_FIRST,
            repo: repoID,
            revspec: commitID,
            filePath,
            afterDate: showOlderCommits ? null : formatISO(subYears(Date.now(), 1)),
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data.node || data.node.__typename !== 'Repository') {
                throw new Error('Unable to get repository')
            }
            if (!data.node.commit) {
                throw new Error('Unable to get commit')
            }
            return data.node.commit.ancestors
        },
    })

    const onShowMoreClick = useCallback((): void => {
        setShowOlderCommits(true)
        fetchMore()
    }, [fetchMore])

    return (
        <ConnectionContainer className={className} compact={true}>
            {error && <ConnectionError errors={[error.message]} compact={true} />}
            {connection &&
                (connection.nodes.length > 0 ? (
                    <ConnectionList compact={true} className="list-group list-group-flush">
                        {connection.nodes.map(node => (
                            <GitCommitNode
                                key={node.id}
                                node={node}
                                className={classNames('list-group-item', styles.gitCommitNode)}
                                messageSubjectClassName={styles.gitCommitNodeMessageSubject}
                                compact={true}
                                tag="li"
                            />
                        ))}
                    </ConnectionList>
                ) : showOlderCommits ? (
                    <>No commits in this tree.</>
                ) : (
                    <div className="test-tree-page-no-recent-commits">
                        <p className="mb-2">No commits in the last year.</p>
                        <button
                            type="button"
                            className="btn btn-secondary btn-sm test-tree-page-show-all-commits"
                            onClick={onShowOlderCommitsClick}
                        >
                            Show older commits
                        </button>
                    </div>
                ))}
            {loading && <ConnectionLoading compact={true} />}
            {!loading && connection && connection.nodes.length > 0 && (
                <SummaryContainer compact={true}>
                    {(hasNextPage || !showOlderCommits) && <ShowMoreButton compact={true} onClick={onShowMoreClick} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}
