import React, { useState, useCallback } from 'react'

import classNames from 'classnames'
import { subYears, formatISO } from 'date-fns'
import * as H from 'history'
import { Observable } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, pluralize, encodeURIPathComponent } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import {
    Button,
    Link,
    Badge,
    useEventObservable,
    Alert,
    LoadingSpinner,
    H2,
    Text,
    ButtonLink,
} from '@sourcegraph/wildcard'

import { BatchChangesProps } from '../../batches'
import { CodeIntelligenceProps } from '../../codeintel'
import { FilteredConnection } from '../../components/FilteredConnection'
import {
    GetRepoBatchChangesSummaryResult,
    GetRepoBatchChangesSummaryVariables,
    GitCommitFields,
    TreePageRepositoryFields,
} from '../../graphql-operations'
import { fetchBlob } from '../blob/backend'
import { BlobInfo } from '../blob/Blob'
import { RenderedFile } from '../blob/RenderedFile'
import { GitCommitNode, GitCommitNodeProps } from '../commits/GitCommitNode'

import { fetchTreeCommits } from './TreePageContent'

import styles from './HomeTab.module.scss'

interface Props extends SettingsCascadeProps, CodeIntelligenceProps, BatchChangesProps {
    repo: TreePageRepositoryFields
    filePath: string
    commitID: string
    revision: string
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

export const HomeTab: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repo,
    commitID,
    revision,
    filePath,
    codeIntelligenceEnabled,
    codeIntelligenceBadgeContent: CodeIntelligenceBadge,
    batchChangesEnabled,
    ...props
}) => {
    const [richHTML, setRichHTML] = useState<string | null>('loading')
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
                            revision,
                            filePath: `${filePath}/README.md`,
                            disableTimeout,
                        })
                    ),
                    map(blob => {
                        if (blob === null) {
                            setRichHTML(null)
                            return blob
                        }

                        // Replace html with lsif generated HTML, if available
                        if (blob.richHTML) {
                            setRichHTML(blob.richHTML)
                            setAborted(blob.highlight.aborted || false)
                        } else {
                            setRichHTML(null)
                        }

                        const blobInfo: BlobInfo & { richHTML: string; aborted: boolean } = {
                            content: blob.content,
                            html: blob.highlight.html ?? '',
                            repoName: repo.name,
                            revision,
                            commitID,
                            filePath: `${filePath}/README.md`,
                            mode: '',
                            // Properties used in `BlobPage` but not `Blob`
                            richHTML: blob.richHTML,
                            aborted: blob.highlight.aborted,
                        }
                        return blobInfo
                    }),
                    catchError((error): [ErrorLike] => [asError(error)])
                ),
            [repo.name, commitID, filePath, revision]
        )
    )

    const onExtendTimeoutClick = useCallback(
        (event: React.MouseEvent): void => {
            event.preventDefault()
            nextFetchWithDisabledTimeout()
        },
        [nextFetchWithDisabledTimeout]
    )

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
        <div className="w-100">No commits in this tree.</div>
    ) : (
        <div className="test-tree-page-no-recent-commits w-100">
            <Text className="mb-2">No commits in this tree in the past year.</Text>
            <div className="float-right">
                <Button onClick={onShowOlderCommitsClicked} variant="link" size="sm" className="float-right p-0">
                    Show older commits
                </Button>
            </div>
        </div>
    )

    const TotalCountSummary: React.FunctionComponent<React.PropsWithChildren<{ totalCount: number }>> = ({
        totalCount,
    }) => (
        <div className="mt-2 w-100">
            {showOlderCommits ? (
                <>
                    {totalCount} total {pluralize('commit', totalCount)} in this tree.
                </>
            ) : (
                <>
                    <Text className="mb-2">
                        {totalCount} {pluralize('commit', totalCount)} in this tree in the past year.
                    </Text>
                    <div className="float-right">
                        <Button
                            onClick={onShowOlderCommitsClicked}
                            variant="link"
                            size="sm"
                            className="float-right p-0"
                        >
                            Show all commits
                        </Button>
                    </div>
                </>
            )}
        </div>
    )

    interface RecentCommitsProps {
        isSidebar: boolean
    }

    const RecentCommits: React.FunctionComponent<React.PropsWithChildren<RecentCommitsProps>> = ({ isSidebar }) => (
        <div className="mb-3">
            <H2>Recent commits</H2>
            <FilteredConnection<
                GitCommitFields,
                Pick<
                    GitCommitNodeProps,
                    | 'className'
                    | 'compact'
                    | 'messageSubjectClassName'
                    | 'hideExpandCommitMessageBody'
                    | 'sidebar'
                    | 'wrapperElement'
                >
            >
                location={props.location}
                className="mt-2 p0 m-0"
                listClassName="list-group list-group-flush"
                noun="commit in this tree"
                pluralNoun="commits in this tree"
                queryConnection={queryCommits}
                nodeComponent={GitCommitNode}
                showMoreClassName="px-0"
                nodeComponentProps={{
                    className: classNames('list-group-item', styles.gitCommitNode),
                    messageSubjectClassName: isSidebar ? 'd-none' : styles.gitCommitNodeMessageSubject,
                    compact: isSidebar,
                    hideExpandCommitMessageBody: isSidebar,
                    sidebar: isSidebar,
                    wrapperElement: 'li',
                }}
                updateOnChange={`${repo.name}:${revision}:${filePath}:${String(showOlderCommits)}`}
                defaultFirst={7}
                useURLQuery={false}
                hideSearch={true}
                emptyElement={emptyElement}
                totalCountSummaryComponent={TotalCountSummary}
            />
        </div>
    )

    const READMEFile: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
        <div>
            {richHTML && richHTML !== 'loading' && (
                <RenderedFile className="pt-0 pl-3" dangerousInnerHTML={richHTML} location={props.location} />
            )}
            {!richHTML && richHTML !== 'loading' && (
                <div className="text-center mt-5">
                    <img src="https://i.ibb.co/tztztYB/eric.png" alt="winner" className="mb-3 w-25" />
                    <H2>No README available :)</H2>
                </div>
            )}
            {blobInfoOrError && richHTML && aborted && (
                <div>
                    <Alert variant="info">
                        Rendering this file took too long. &nbsp;
                        <Button onClick={onExtendTimeoutClick} variant="primary" size="sm">
                            Try again
                        </Button>
                    </Alert>
                </div>
            )}
        </div>
    )

    // Only render recent commits and readme for non-root directory
    if (filePath) {
        return (
            <div className="container mw-100">
                <RecentCommits isSidebar={false} />
                <H2 className="mt-5">README.md</H2>
                <READMEFile />
            </div>
        )
    }

    return (
        <div className="container mw-100">
            <div className="row">
                {/* RENDER README */}
                <div className="col-sm m-0 pl-0 pt-0">
                    <READMEFile />
                </div>
                {/* SIDE MENU*/}
                <div className="col-sm col-lg-4 m-0">
                    <div className={styles.section}>
                        <RecentCommits isSidebar={true} />
                        {/* CODE-INTEL */}
                        <div className="mb-3">
                            <H2>Code intel</H2>
                            {CodeIntelligenceBadge && (
                                <CodeIntelligenceBadge
                                    repoName={repo.name}
                                    revision={revision}
                                    filePath={filePath}
                                    {...props}
                                />
                            )}
                        </div>
                        {/* BATCH CHANGES */}
                        <div className="mb-3">
                            <H2>Batch changes</H2>
                            {batchChangesEnabled ? (
                                <HomeTabBatchChangeBadge repoName={repo.name} />
                            ) : (
                                <div>
                                    <div className={styles.item}>
                                        <Badge
                                            variant="danger"
                                            className={classNames('text-uppercase col-4', styles.itemBadge)}
                                        >
                                            DISABLED
                                        </Badge>
                                        <div className="col">Not available</div>
                                    </div>
                                    <div className="text-right">
                                        <ButtonLink
                                            size="sm"
                                            className="mr-0 pr-0"
                                            to="/help/batch_changes"
                                            variant="link"
                                        >
                                            Learn more
                                        </ButtonLink>
                                    </div>
                                </div>
                            )}
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

export const HomeTabBatchChangeBadge: React.FunctionComponent<
    React.PropsWithChildren<HomeTabBatchChangeBadgeProps>
> = ({ repoName }) => {
    const { loading, error, data } = useQuery<GetRepoBatchChangesSummaryResult, GetRepoBatchChangesSummaryVariables>(
        REPO_BATCH_CHANGES_SUMMARY,
        {
            variables: { name: repoName },
        }
    )
    if (loading) {
        return (
            <div className={styles.item}>
                <LoadingSpinner />
            </div>
        )
    }

    const allBatchChanges = (
        <div className="text-right">
            <ButtonLink size="sm" variant="link" to={`/${encodeURIPathComponent(repoName)}/-/batch-changes`}>
                View all batch changes
            </ButtonLink>
        </div>
    )

    if (error || !data?.repository) {
        return (
            <>
                <div className={styles.item}>
                    <ErrorMessage error="Failed to load batch changes" />
                </div>
                {allBatchChanges}
            </>
        )
    }

    const badgeClassNames = classNames('text-uppercase col-2 align-self-center', styles.itemBadge)
    const batchChanges = data?.repository?.batchChanges
    if (!batchChanges || batchChanges.nodes.length === 0) {
        return (
            <>
                <div className={classNames(styles.item, 'flex-column')}>
                    <div className="mb-1">No open batch changes for this repository</div>
                    <Link to="/batch-changes/create">Create a batch change</Link>
                </div>
                {allBatchChanges}
            </>
        )
    }

    const items: React.ReactElement[] = batchChanges.nodes.map(
        ({ id, name, namespace: { namespaceName }, changesetsStats, url }) => {
            const summaries: { value: number; name: string }[] = [
                {
                    name: 'open',
                    value: changesetsStats.open,
                },
                {
                    name: 'merged',
                    value: changesetsStats.merged,
                },
                {
                    name: 'closed',
                    value: changesetsStats.closed,
                },
            ]
            const summaryTexts = summaries.map(({ value, name }) => `${value} ${name}`)

            return (
                <li className={styles.item} key={id}>
                    <Badge variant="success" className={badgeClassNames}>
                        OPEN
                    </Badge>
                    <div className={classNames('d-block col', styles.itemBatchChangeText)}>
                        <Link to={url}>
                            {namespaceName} / {name}
                        </Link>
                        <div>{summaryTexts.join(', ')}</div>
                    </div>
                </li>
            )
        }
    )
    return (
        <>
            <ul className={styles.list}>{items}</ul>
            {allBatchChanges}
        </>
    )
}

const REPO_BATCH_CHANGE_FRAGMENT = gql`
    fragment RepoBatchChangeSummary on BatchChange {
        id
        state
        name
        namespace {
            namespaceName
        }
        url
        changesetsStats {
            open
            merged
            closed
        }
    }
`

const REPO_BATCH_CHANGES_SUMMARY = gql`
    query GetRepoBatchChangesSummary($name: String!) {
        repository(name: $name) {
            batchChanges(state: OPEN, first: 10) {
                nodes {
                    ...RepoBatchChangeSummary
                }
            }
        }
    }

    ${REPO_BATCH_CHANGE_FRAGMENT}
`
