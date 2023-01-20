import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import { formatISO, subYears } from 'date-fns'
import * as H from 'history'
import { escapeRegExp } from 'lodash'
import { Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { memoizeObservable, numberWithCommas, pluralize } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SearchPatternType, TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Card, CardHeader, Link, Tooltip, Text } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
    ConnectionError,
} from '../../components/FilteredConnection/ui'
import {
    BlobFileFields,
    CommitAtTimeResult,
    CommitAtTimeVariables,
    DiffSinceResult,
    DiffSinceVariables,
    GitCommitFields,
    RepositoryContributorNodeFields,
    TreePageRepositoryContributorsResult,
    TreePageRepositoryContributorsVariables,
    Scalars,
    TreeCommitsResult,
    TreePageRepositoryFields,
    TreeCommitsVariables,
} from '../../graphql-operations'
import { PersonLink } from '../../person/PersonLink'
import { quoteIfNeeded, searchQueryForRepoRevision } from '../../search'
import { UserAvatar } from '../../user/UserAvatar'
import { fetchBlob } from '../blob/backend'
import { GitCommitNode, GitCommitNodeProps } from '../commits/GitCommitNode'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'
import { BATCH_COUNT } from '../RepositoriesPopover'

import { DiffStat, FilesCard, ReadmePreviewCard } from './TreePagePanels'

import styles from './TreePageContent.module.scss'
import contributorsStyles from './TreePageContentContributors.module.scss'
import panelStyles from './TreePagePanels.module.scss'

export type TreeCommitsRepositoryCommit = NonNullable<
    Extract<TreeCommitsResult['node'], { __typename: 'Repository' }>['commit']
>

export const fetchTreeCommits = memoizeObservable(
    (args: {
        repo: Scalars['ID']
        revspec: string
        first?: number
        filePath?: string
        after?: string
    }): Observable<TreeCommitsRepositoryCommit['ancestors']> =>
        requestGraphQL<TreeCommitsResult, TreeCommitsVariables>(
            gql`
                query TreeCommits($repo: ID!, $revspec: String!, $first: Int, $filePath: String, $after: String) {
                    node(id: $repo) {
                        __typename
                        ... on Repository {
                            commit(rev: $revspec) {
                                ancestors(first: $first, path: $filePath, after: $after) {
                                    nodes {
                                        ...GitCommitFields
                                    }
                                    pageInfo {
                                        hasNextPage
                                    }
                                }
                            }
                        }
                    }
                }
                ${gitCommitFragment}
            `,
            {
                ...args,
                first: args.first || null,
                filePath: args.filePath || null,
                after: args.after || null,
            }
        ).pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.node) {
                    throw new Error('Repository not found')
                }
                if (data.node.__typename !== 'Repository') {
                    throw new Error('Node is not a Repository')
                }
                if (!data.node.commit) {
                    throw new Error('Commit not found')
                }
                return data.node.commit.ancestors
            })
        ),
    args => `${args.repo}:${args.revspec}:${String(args.first)}:${String(args.filePath)}:${String(args.after)}`
)

export const fetchCommit = (args: {
    repo: Scalars['String']
    revspec: Scalars['String']
    beforespec: Scalars['String'] | null
}): Observable<GitCommitFields> =>
    requestGraphQL<CommitAtTimeResult, CommitAtTimeVariables>(
        gql`
            query CommitAtTime($repo: String!, $revspec: String!, $beforespec: String) {
                repository(name: $repo) {
                    commit(rev: $revspec) {
                        ancestors(first: 1, before: $beforespec) {
                            nodes {
                                ...GitCommitFields
                            }
                        }
                    }
                }
            }
            ${gitCommitFragment}
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            const nodes = data.repository?.commit?.ancestors.nodes
            if (!nodes || nodes.length === 0) {
                throw new Error(`no commit found before ${args.beforespec} from revspec ${args.revspec}`)
            }
            return nodes[0]
        })
    )

export const fetchDiffStats = (args: {
    repo: Scalars['String']
    revspec: Scalars['String']
    beforespec: Scalars['String']
    filePath: Scalars['String']
}): Observable<DiffStat[]> =>
    fetchCommit({
        repo: args.repo,
        revspec: args.revspec,
        beforespec: null,
    }).pipe(
        switchMap(headCommit => {
            const headDate = new Date(Date.parse(headCommit.author.date))
            const absBeforespec = `${headDate.getUTCFullYear()}-${
                headDate.getUTCMonth() + 1
            }-${headDate.getUTCDate()} ${args.beforespec}`
            return fetchCommit({
                repo: args.repo,
                revspec: args.revspec,
                beforespec: absBeforespec,
            })
        }),
        switchMap((base: GitCommitFields) =>
            requestGraphQL<DiffSinceResult, DiffSinceVariables>(
                gql`
                    query DiffSince($repo: String!, $basespec: String!, $headspec: String!, $filePath: String!) {
                        repository(name: $repo) {
                            comparison(base: $basespec, head: $headspec) {
                                fileDiffs(paths: [$filePath]) {
                                    nodes {
                                        newPath
                                        stat {
                                            added
                                            deleted
                                        }
                                    }
                                }
                            }
                        }
                    }
                `,
                {
                    repo: args.repo,
                    basespec: base.oid,
                    headspec: args.revspec,
                    filePath: args.filePath,
                }
            )
        ),
        map(dataOrThrowErrors),
        map(
            (data): DiffStat[] =>
                data.repository?.comparison.fileDiffs.nodes
                    ?.filter(node => node.newPath)
                    .map(node => ({
                        path: node.newPath!,
                        ...node.stat,
                    })) || []
        ),
        map((fileDiffStats: DiffStat[]) => {
            const aggregatedDiffStats: { [path: string]: DiffStat } = {}
            for (const diffStat of fileDiffStats) {
                // strip filePath prefix from fileDiffStat.path
                const strippedPath =
                    args.filePath === '.' ? diffStat.path : diffStat.path.slice(args.filePath.length + 1)
                let subdirName = strippedPath
                if (subdirName.includes('/')) {
                    subdirName = subdirName.slice(0, subdirName.indexOf('/'))
                }
                if (!aggregatedDiffStats[subdirName]) {
                    aggregatedDiffStats[subdirName] = { path: subdirName, added: 0, deleted: 0 }
                }
                aggregatedDiffStats[subdirName].added += diffStat.added
                aggregatedDiffStats[subdirName].deleted += diffStat.deleted
            }
            return Array.from(Object.values(aggregatedDiffStats))
        })
    )

interface TreePageContentProps extends ExtensionsControllerProps, ThemeProps, TelemetryProps, PlatformContextProps {
    filePath: string
    tree: TreeFields
    repo: TreePageRepositoryFields
    commitID: string
    location: H.Location
    revision: string
}

export const TreePageContent: React.FunctionComponent<React.PropsWithChildren<TreePageContentProps>> = ({
    filePath,
    tree,
    repo,
    commitID,
    revision,
    ...props
}) => {
    const [showOlderCommits, setShowOlderCommits] = useState(false)
    const [readmeInfo, setReadmeInfo] = useState<
        | undefined
        | {
              blob: BlobFileFields
              entry: TreeFields['entries'][number]
          }
    >()
    useEffect(() => {
        const readmeEntry = (() => {
            for (const readmeName of ['README.md', 'README']) {
                for (const entry of tree.entries) {
                    if (!entry.isDirectory && entry.name === readmeName) {
                        return entry
                    }
                }
            }
            return null
        })()
        if (!readmeEntry) {
            setReadmeInfo(undefined)
            return
        }

        const subscription = fetchBlob({
            repoName: repo.name,
            revision,
            filePath: readmeEntry?.path,
            disableTimeout: true,
        }).subscribe(blob => {
            if (blob) {
                setReadmeInfo({
                    blob,
                    entry: readmeEntry,
                })
            } else {
                setReadmeInfo(undefined)
            }
        })
        return () => subscription.unsubscribe()
    }, [repo.name, revision, filePath, tree.entries])

    const [diffStats, setDiffStats] = useState<DiffStat[]>()
    useEffect(() => {
        const subscription = fetchDiffStats({
            repo: repo.name,
            revspec: revision,
            beforespec: '1 month',
            filePath: filePath || '.',
        }).subscribe(results => {
            setDiffStats(results)
        })
        return () => subscription.unsubscribe()
    }, [repo.name, revision, filePath])

    const queryCommits = useCallback(
        (args: { first?: number }): Observable<TreeCommitsRepositoryCommit['ancestors']> => {
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

    const onShowOlderCommitsClicked = useCallback((event: React.MouseEvent): void => {
        event.preventDefault()
        setShowOlderCommits(true)
    }, [])

    const emptyElement = showOlderCommits ? (
        <>No commits in this tree.</>
    ) : (
        <div className="test-tree-page-no-recent-commits">
            <Text className="mb-2">No commits in this tree in the past year.</Text>
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

    const TotalCountSummary: React.FunctionComponent<React.PropsWithChildren<{ totalCount: number }>> = ({
        totalCount,
    }) => (
        <div className="p-2">
            {showOlderCommits ? (
                <>
                    {totalCount} total {pluralize('commit', totalCount)} in this tree.
                </>
            ) : (
                <>
                    <Text className="mb-2">
                        {totalCount} {pluralize('commit', totalCount)} in this tree in the past year.
                    </Text>
                    <Button onClick={onShowOlderCommitsClicked} variant="secondary" size="sm">
                        Show all commits
                    </Button>
                </>
            )}
        </div>
    )

    return (
        <>
            {readmeInfo && (
                <ReadmePreviewCard
                    readmeHTML={readmeInfo.blob.richHTML}
                    readmeURL={readmeInfo.entry.url}
                    location={props.location}
                    className="mb-4"
                />
            )}
            <section className={classNames('test-tree-entries container mb-3 px-0', styles.section)}>
                <FilesCard diffStats={diffStats} entries={tree.entries} className={styles.files} filePath={filePath} />

                <Card className={styles.commits}>
                    <CardHeader className={panelStyles.cardColHeaderWrapper}>
                        {tree.isRoot ? <Link to={`${tree.url}/-/commits`}>Commits</Link> : 'Commits'}
                    </CardHeader>

                    <FilteredConnection<
                        GitCommitFields,
                        Pick<GitCommitNodeProps, 'className' | 'compact' | 'messageSubjectClassName' | 'wrapperElement'>
                    >
                        location={props.location}
                        listClassName="list-group list-group-flush"
                        noun="commit in this tree"
                        pluralNoun="commits in this tree"
                        queryConnection={queryCommits}
                        nodeComponent={GitCommitNode}
                        nodeComponentProps={{
                            className: classNames('list-group-item px-2 py-1', styles.gitCommitNode),
                            messageSubjectClassName: styles.gitCommitNodeMessageSubject,
                            compact: true,
                            wrapperElement: 'li',
                        }}
                        updateOnChange={`${repo.name}:${revision}:${filePath}:${String(showOlderCommits)}`}
                        defaultFirst={20}
                        useURLQuery={false}
                        hideSearch={true}
                        emptyElement={emptyElement}
                        totalCountSummaryComponent={TotalCountSummary}
                        loaderClassName={contributorsStyles.filteredConnectionLoading}
                        showMoreClassName="mb-0"
                        summaryClassName={contributorsStyles.filteredConnectionSummary}
                    />
                </Card>

                <Card className={styles.contributors}>
                    <CardHeader className={panelStyles.cardColHeaderWrapper}>
                        {tree.isRoot ? (
                            <Link to={`${tree.url}/-/stats/contributors`}>Contributors</Link>
                        ) : (
                            'Contributors'
                        )}
                    </CardHeader>
                    <Contributors
                        filePath={filePath}
                        tree={tree}
                        repo={repo}
                        commitID={commitID}
                        revision={revision}
                        {...props}
                    />
                </Card>
            </section>
        </>
    )
}

const CONTRIBUTORS_QUERY = gql`
    query TreePageRepositoryContributors(
        $repo: ID!
        $first: Int
        $revisionRange: String
        $afterDate: String
        $path: String
    ) {
        node(id: $repo) {
            ... on Repository {
                contributors(first: $first, revisionRange: $revisionRange, afterDate: $afterDate, path: $path) {
                    ...TreePageRepositoryContributorConnectionFields
                }
            }
        }
    }

    fragment TreePageRepositoryContributorConnectionFields on RepositoryContributorConnection {
        totalCount
        pageInfo {
            hasNextPage
        }
        nodes {
            ...TreePageRepositoryContributorNodeFields
        }
    }

    fragment TreePageRepositoryContributorNodeFields on RepositoryContributor {
        person {
            name
            displayName
            email
            avatarURL
            user {
                username
                url
                displayName
                avatarURL
            }
        }
        count
        commits(first: 1) {
            nodes {
                oid
                abbreviatedOID
                url
                subject
                author {
                    date
                }
            }
        }
    }
`

interface ContributorsProps extends TreePageContentProps {}

const Contributors: React.FunctionComponent<ContributorsProps> = ({ repo, filePath }) => {
    const spec: QuerySpec = {
        revisionRange: '',
        after: '',
        path: filePath,
    }

    const { connection, error, loading, hasNextPage, fetchMore } = useShowMorePagination<
        TreePageRepositoryContributorsResult,
        TreePageRepositoryContributorsVariables,
        RepositoryContributorNodeFields
    >({
        query: CONTRIBUTORS_QUERY,
        variables: {
            first: BATCH_COUNT,
            repo: repo.id,
            revisionRange: spec.revisionRange,
            afterDate: spec.after,
            path: filePath,
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)
            if (!node) {
                throw new Error(`Node ${repo.id} not found`)
            }
            if (!('contributors' in node)) {
                throw new Error('Failed to fetch contributors for this repo')
            }
            return node.contributors
        },
        options: {
            fetchPolicy: 'cache-first',
        },
    })

    return (
        <ConnectionContainer>
            {error && <ConnectionError errors={[error.message]} />}
            {connection && connection.nodes.length > 0 && (
                <ConnectionList className="list-group list-group-flush test-filtered-contributors-connection">
                    {connection.nodes.map(node => (
                        <RepositoryContributorNode
                            key={`${node.person.displayName}:${node.count}`}
                            node={node}
                            repoName={repo.name}
                            // TODO: what does `globbing` do?
                            globbing={true}
                            {...spec}
                        />
                    ))}
                </ConnectionList>
            )}
            {loading && (
                <div className={contributorsStyles.filteredConnectionLoading}>
                    <ConnectionLoading />
                </div>
            )}
            <SummaryContainer className={styles.contributorsSummary}>
                {connection && (
                    <ConnectionSummary
                        compact={true}
                        connection={connection}
                        first={BATCH_COUNT}
                        noun="contributor"
                        pluralNoun="contributors"
                        hasNextPage={hasNextPage}
                    />
                )}
                {hasNextPage && <ShowMoreButton className="m-0 p-1 border-0" onClick={fetchMore} />}
            </SummaryContainer>
        </ConnectionContainer>
    )
}

interface QuerySpec {
    revisionRange: string
    after: string
    path: string
}

interface RepositoryContributorNodeProps extends QuerySpec {
    node: RepositoryContributorNodeFields
    repoName: string
    globbing: boolean
}

const RepositoryContributorNode: React.FunctionComponent<React.PropsWithChildren<RepositoryContributorNodeProps>> = ({
    node,
    repoName,
    revisionRange,
    after,
    path,
    globbing,
}) => {
    const query: string = [
        searchQueryForRepoRevision(repoName, globbing),
        'type:diff',
        `author:${quoteIfNeeded(node.person.email)}`,
        after ? `after:${quoteIfNeeded(after)}` : '',
        path ? `file:${quoteIfNeeded(escapeRegExp(path))}` : '',
    ]
        .join(' ')
        .replace(/\s+/, ' ')

    return (
        <li className={classNames('list-group-item py-2', contributorsStyles.repositoryContributorNode)}>
            <div className={contributorsStyles.person}>
                <UserAvatar inline={true} className="mr-2" user={node.person.user ? node.person.user : node.person} />
                <PersonLink userClassName="font-weight-bold" person={node.person} />
            </div>
            <div className={contributorsStyles.commits}>
                <Tooltip
                    content={
                        revisionRange?.includes('..')
                            ? 'All commits will be shown (revision end ranges are not yet supported)'
                            : null
                    }
                    placement="left"
                >
                    <Link
                        to={`/search?${buildSearchURLQuery(query, SearchPatternType.standard, false)}`}
                        className="font-weight-bold"
                    >
                        {numberWithCommas(node.count)} {pluralize('commit', node.count)}
                    </Link>
                </Tooltip>
            </div>
        </li>
    )
}
