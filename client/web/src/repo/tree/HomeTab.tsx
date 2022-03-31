import React, { useState, useMemo, useCallback, useEffect } from 'react'

import classNames from 'classnames'
import { subYears, formatISO } from 'date-fns'
import * as H from 'history'
import { Observable } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'

import { asError, ErrorLike, pluralize, encodeURIPathComponent } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, useObservable, Link, Badge, useEventObservable, Alert } from '@sourcegraph/wildcard'

import { FilteredConnection } from '../../components/FilteredConnection'
import { queryRepoBatchChangeStats } from '../../enterprise/batches/repo/backend'
import { GitCommitFields, RepoBatchChangeStats, TreePageRepositoryFields } from '../../graphql-operations'
import { fetchBlob } from '../blob/backend'
import { BlobInfo } from '../blob/Blob'
import { RenderedFile } from '../blob/RenderedFile'
import { GitCommitNode, GitCommitNodeProps } from '../commits/GitCommitNode'

import { fetchTreeCommits } from './TreePageContent'

import styles from './HomeTab.module.scss'

interface Props {
    repo: TreePageRepositoryFields
    filePath: string
    commitID: string
    revision: string
    codeIntelligenceEnabled: boolean
    batchChangesEnabled: boolean
    location: H.Location
    history?: H.History
    globbing?: boolean
}

export const treePageRepositoryFragment = gql`
    fragment TreePageRepositoryFields on Repository {
        id
        name
        description
        viewerCanAdminister
        url
    }
`

export const HomeTab: React.FunctionComponent<Props> = ({
    repo,
    commitID,
    revision,
    filePath,
    codeIntelligenceEnabled,
    batchChangesEnabled,
    ...props
}) => {
    const [richHTML, setRichHTML] = useState('')
    const [aborted, setAborted] = useState(false)
    const [nextFetchWithDisabledTimeout, blobInfoOrError] = useEventObservable<
        void,
        (BlobInfo & { richHTML: string; aborted: boolean }) | null | ErrorLike
    >(
        useCallback(
            (clicks: Observable<void>) =>
                clicks.pipe(
                    mapTo(true),
                    startWith(false),
                    switchMap(disableTimeout =>
                        fetchBlob({
                            repoName: repo.name,
                            commitID,
                            filePath: 'README.md',
                            disableTimeout,
                        })
                    ),
                    map(blob => {
                        if (blob === null) {
                            return blob
                        }

                        // Replace html with lsif generated HTML, if available
                        if (blob.richHTML) {
                            setRichHTML(blob.richHTML)
                            setAborted(blob.highlight.aborted)
                        }

                        const blobInfo: BlobInfo & { richHTML: string; aborted: boolean } = {
                            content: blob.content,
                            html: blob.highlight.html,
                            repoName: repo.name,
                            revision,
                            commitID,
                            filePath: 'README.md',
                            mode: '',
                            // Properties used in `BlobPage` but not `Blob`
                            richHTML: blob.richHTML,
                            aborted: blob.highlight.aborted,
                        }
                        return blobInfo
                    }),
                    catchError((error): [ErrorLike] => [asError(error)])
                ),
            [repo.name, commitID, revision]
        )
    )

    const onExtendTimeoutClick = useCallback(
        (event: React.MouseEvent): void => {
            event.preventDefault()
            nextFetchWithDisabledTimeout()
        },
        [nextFetchWithDisabledTimeout]
    )

    useEffect(() => {
        if (!blobInfoOrError) {
            console.error('error')
        }
    }, [blobInfoOrError])

    const [showOlderCommits, setShowOlderCommits] = useState(false)

    const onShowOlderCommitsClicked = useCallback(
        (event: React.MouseEvent): void => {
            event.preventDefault()
            setShowOlderCommits(true)
        },
        [setShowOlderCommits]
    )

    const queryCommits = useCallback(
        (args: { first?: number }): Observable<GQL.IGitCommitConnection> => {
            const after: string | undefined = showOlderCommits ? undefined : formatISO(subYears(Date.now(), 1))
            return fetchTreeCommits({
                ...args,
                repo: repo.id,
                revspec: revision || '',
                filePath,
                after,
            })
        },
        [filePath, repo.id, revision, showOlderCommits]
    )

    const emptyElement = showOlderCommits ? (
        <>No commits in this tree.</>
    ) : (
        <div className="test-tree-page-no-recent-commits">
            <p className="mb-2">No commits in this tree in the past year.</p>
            <Button
                className="test-tree-page-show-all-commits"
                onClick={onShowOlderCommitsClicked}
                variant="secondary"
                size="sm"
            >
                Show all commits
            </Button>
        </div>
    )

    const TotalCountSummary: React.FunctionComponent<{ totalCount: number }> = ({ totalCount }) => (
        <div className="mt-2">
            {showOlderCommits ? (
                <>
                    {totalCount} total {pluralize('commit', totalCount)} in this tree.
                </>
            ) : (
                <>
                    <p className="mb-2">
                        {totalCount} {pluralize('commit', totalCount)} in this tree in the past year.
                    </p>
                    <Button onClick={onShowOlderCommitsClicked} variant="secondary" size="sm">
                        Show all commits
                    </Button>
                </>
            )}
        </div>
    )

    return (
        <div className="container p-0 m-0 mw-100">
            <div className="row">
                <div className="col-sm m-0">
                    {richHTML && <RenderedFile dangerousInnerHTML={richHTML} location={props.location} />}
                    {blobInfoOrError && richHTML && aborted && (
                        <div>
                            <Alert variant="info">
                                Syntax-highlighting this file took too long. &nbsp;
                                <Button onClick={onExtendTimeoutClick} variant="primary" size="sm">
                                    Try again
                                </Button>
                            </Alert>
                        </div>
                    )}
                </div>
                <div className="col-sm col-lg-4 m-0">
                    <div className="mb-5">
                        <div className={styles.section}>
                            <h2>Code Intel</h2>
                            <div className={styles.item}>
                                <Badge
                                    variant={codeIntelligenceEnabled ? 'primary' : 'danger'}
                                    className={classNames('text-uppercase col-4')}
                                >
                                    {codeIntelligenceEnabled ? 'AVAILABLE' : 'DISABLED'}
                                </Badge>
                                <div className="d-block col">
                                    <div>Precise code intelligence</div>
                                </div>
                            </div>
                        </div>
                        <div className={styles.section}>
                            <h2>Batch Changes</h2>
                            {batchChangesEnabled ? (
                                <HomeTabBatchChangeBadge repoName={repo.name} />
                            ) : (
                                <div className={styles.item}>
                                    <Badge variant="danger" className={classNames('text-uppercase col-4')}>
                                        DISABLED
                                    </Badge>
                                    <div className="col">Not available</div>
                                </div>
                            )}
                        </div>
                        <div className={styles.section}>
                            <h2>Recent Commits</h2>
                            <FilteredConnection<
                                GitCommitFields,
                                Pick<GitCommitNodeProps, 'className' | 'compact' | 'messageSubjectClassName'>
                            >
                                location={props.location}
                                className="mt-2"
                                listClassName="list-group list-group-flush"
                                noun="commit in this tree"
                                pluralNoun="commits in this tree"
                                queryConnection={queryCommits}
                                nodeComponent={GitCommitNode}
                                nodeComponentProps={{
                                    className: classNames('list-group-item', styles.gitCommitNode),
                                    messageSubjectClassName: undefined,
                                    compact: true,
                                }}
                                updateOnChange={`${repo.name}:${revision}:${filePath}:${String(showOlderCommits)}`}
                                defaultFirst={7}
                                useURLQuery={false}
                                hideSearch={true}
                                emptyElement={emptyElement}
                                totalCountSummaryComponent={TotalCountSummary}
                            />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

interface HomeTabBatchChangeBadgeProps {
    repoName: string
}

export const HomeTabBatchChangeBadge: React.FunctionComponent<HomeTabBatchChangeBadgeProps> = ({ repoName }) => {
    const stats: RepoBatchChangeStats | undefined = useObservable(
        useMemo(() => queryRepoBatchChangeStats({ name: repoName }), [repoName])
    )
    const hasChangesets = stats?.changesetsStats.total
    return (
        <>
            {hasChangesets && hasChangesets > 0 && stats?.batchChangesDiffStat && stats?.changesetsStats ? (
                <>
                    {stats?.changesetsStats.open > 0 && (
                        <div className={styles.item}>
                            <Badge variant="success" className={classNames('text-uppercase col-4')}>
                                OPEN
                            </Badge>
                            <div className="col">{stats?.changesetsStats.open} open changesets</div>
                        </div>
                    )}
                    {stats?.changesetsStats.unpublished > 0 && (
                        <div className={styles.item}>
                            <Badge variant="secondary" className={classNames('text-uppercase col-4')}>
                                Unpublished
                            </Badge>
                            <div className="col">{stats?.changesetsStats.unpublished} unpublished changesets</div>
                        </div>
                    )}
                    {stats?.changesetsStats.draft > 0 && (
                        <div className={styles.item}>
                            <Badge variant="secondary" className={classNames('text-uppercase col-4')}>
                                draft
                            </Badge>
                            <div className="col">{stats?.changesetsStats.draft} draft changesets</div>
                        </div>
                    )}
                    {stats?.changesetsStats.closed > 0 && (
                        <div className={styles.item}>
                            <Badge variant="secondary" className={classNames('text-uppercase col-4')}>
                                CLOSE
                            </Badge>
                            <div className="col">{stats?.changesetsStats.closed} closed changesets</div>
                        </div>
                    )}
                    <div className="text-right">
                        <Link to={`/${encodeURIPathComponent(repoName)}/-/batch-changes`}>All batch changes</Link>
                    </div>
                </>
            ) : (
                <>
                    <div className={styles.item}>
                        <Badge variant="secondary" className={classNames('text-uppercase col-4')}>
                            Unavailable
                        </Badge>
                        <div className="col">No changeset available</div>
                    </div>
                    <div className="text-right">
                        <Link
                            className="btn btn-sm btn-link"
                            to={`/${encodeURIPathComponent(repoName)}/-/batch-changes`}
                        >
                            Create batch change
                        </Link>
                    </div>
                </>
            )}
        </>
    )
}
