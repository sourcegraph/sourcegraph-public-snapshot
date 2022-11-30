import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import formatISO from 'date-fns/formatISO'
import subYears from 'date-fns/subYears'
import * as H from 'history'
import { escapeRegExp, zip as _zip } from 'lodash'
import { from, Observable, zip } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ContributableMenu } from '@sourcegraph/client-api'
import { numberWithCommas, pluralize } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ActionItem } from '@sourcegraph/shared/src/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/shared/src/actions/ActionsContainer'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SearchPatternType, TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Heading, Link, Tooltip } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import { useConnection } from '../../components/FilteredConnection/hooks/useConnection'
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
    CommitAtTime2Result,
    CommitAtTime2Variables,
    DiffSinceResult,
    DiffSinceVariables,
    GitCommitFields,
    LangStatsResult,
    LangStatsVariables,
    RepositoryContributorNodeFields,
    RepositoryContributorsResult,
    RepositoryContributorsVariables,
    Scalars,
    TreeCommitsResult,
    TreeCommitsVariables,
    TreePageRepositoryFields,
} from '../../graphql-operations'
import { PersonLink } from '../../person/PersonLink'
import { quoteIfNeeded, searchQueryForRepoRevision } from '../../search'
import { UserAvatar } from '../../user/UserAvatar'
import { fetchBlob } from '../blob/backend'
import { GitCommitNode } from '../commits/GitCommitNode'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'
import { BATCH_COUNT } from '../RepositoriesPopover'

import { TreeEntriesSection } from './TreeEntriesSection'
import { DiffStat, LangStats } from './TreePagePanels'

import styles from './TreePage.module.scss'
import contributorsStyles from './TreePageContentContributors.module.scss'

const TREE_COMMITS_PER_PAGE = 10

// TODO(beyang): dark theme, responsive
const TREE_COMMITS_QUERY = gql`
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
`

export const fetchCommit = (args: {
    repo: Scalars['String']
    revspec: Scalars['String']
    beforespec: Scalars['String'] | null
}): Observable<GitCommitFields> =>
    requestGraphQL<CommitAtTime2Result, CommitAtTime2Variables>(
        gql`
            query CommitAtTime2($repo: String!, $revspec: String!, $beforespec: String) {
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

const fetchLangStats = (args: {
    repo: Scalars['String']
    revspec: Scalars['String']
    paths: Scalars['String'][]
}): Observable<LangStats[]> => {
    const langStatsForAllFilesObs = requestGraphQL<LangStatsResult, LangStatsVariables>(
        gql`
            query LangStats($repo: String!, $revspec: String!, $paths: [String!]!) {
                repository(name: $repo) {
                    commit(rev: $revspec) {
                        languageStatistics(paths: $paths) {
                            name
                            totalBytes
                            totalLines
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repository?.commit?.languageStatistics)
    )

    const languageMap = from(import('linguist-languages')).pipe(
        map(({ default: languagesMap }) => (language: string): string => {
            const isLinguistLanguage = (language: string): language is keyof typeof languagesMap =>
                Object.prototype.hasOwnProperty.call(languagesMap, language)

            if (isLinguistLanguage(language)) {
                return languagesMap[language].color ?? 'gray'
            }

            return 'gray'
        })
    )

    return zip(langStatsForAllFilesObs, languageMap).pipe(
        map(([langStatsForAllFiles, getLangColor]) => {
            if (!langStatsForAllFiles) {
                return []
            }

            if (langStatsForAllFiles.length !== args.paths.length) {
                throw new Error('length of language statistics did not match length of entries')
            }

            return _zip(langStatsForAllFiles, args.paths).map(
                ([langStatsForFile, path]): LangStats => ({
                    path: path!,
                    languages: langStatsForFile.map(langStat => ({
                        color: getLangColor(langStat.name),
                        bytes: langStat.totalBytes,
                        lines: langStat.totalLines,
                        name: langStat.name,
                    })),
                })
            )
        })
    )
}

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
    const after = useMemo(() => (showOlderCommits ? null : formatISO(subYears(Date.now(), 1))), [showOlderCommits])

    const { connection, error, loading, hasNextPage, fetchMore, refetchAll } = useShowMorePagination<
        TreeCommitsResult,
        TreeCommitsVariables,
        GitCommitFields
    >({
        query: TREE_COMMITS_QUERY,
        variables: {
            repo: repo.id,
            revspec: revision || '',
            first: TREE_COMMITS_PER_PAGE,
            filePath,
            after,
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                return { nodes: [] }
            }
            if (node.__typename !== 'Repository') {
                return { nodes: [] }
            }
            if (!node.commit?.ancestors) {
                return { nodes: [] }
            }

            return node.commit.ancestors
        },
        options: {
            fetchPolicy: 'cache-first',
        },
    })

    // We store the refetchAll callback in a ref since it will update when
    // variables or result length change and we need to call an up-to-date
    // version in the useEffect below to refetch the proper results.
    //
    // TODO: See if we can make refetchAll stable
    const refetchAllRef = useRef(refetchAll)
    useEffect(() => {
        refetchAllRef.current = refetchAll
    }, [refetchAll])

    useEffect(() => {
        if (showOlderCommits && refetchAllRef.current) {
            // Updating the variables alone is not enough to force a loading
            // indicator to show, so we need to refetch the results.
            refetchAllRef.current()
        }
    }, [showOlderCommits])

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

    const [langStats, setLangStats] = useState<LangStats[]>()
    useEffect(() => {
        const subscription = fetchLangStats({
            repo: repo.name,
            revspec: revision,
            paths: tree.entries.map(entry => entry.path),
        }).subscribe(results => {
            setLangStats(results)
        })
        return () => subscription.unsubscribe()
    }, [repo.name, revision, tree.entries])

    const onShowOlderCommitsClicked = useCallback((event: React.MouseEvent): void => {
        event.preventDefault()
        setShowOlderCommits(true)
    }, [])

    const showAllCommits = (
        <Button
            className="test-tree-page-show-all-commits"
            onClick={onShowOlderCommitsClicked}
            variant="secondary"
            size="sm"
        >
            Show commits older than one year
        </Button>
    )

    // <<<<<<< HEAD
    const { extensionsController } = props

    const showLinkToCommitsPage = connection && hasNextPage && connection.nodes.length > TREE_COMMITS_PER_PAGE

    return (
        <>
            <section className={classNames('test-tree-entries mb-3', styles.section)}>
                <Heading as="h3" styleAs="h2">
                    Files and directories
                </Heading>
                <TreeEntriesSection
                    parentPath={filePath}
                    entries={tree.entries}
                    fileDecorationsByPath={fileDecorationsByPath}
                    isLightTheme={props.isLightTheme}
                />
            </section>
            {extensionsController !== null && window.context.enableLegacyExtensions ? (
                <ActionsContainer
                    {...props}
                    extensionsController={extensionsController}
                    menu={ContributableMenu.DirectoryPage}
                    empty={null}
                >
                    {items => (
                        <section className={styles.section}>
                            <Heading as="h3" styleAs="h2">
                                Actions
                            </Heading>
                            {items.map(item => (
                                <Button
                                    {...props}
                                    extensionsController={extensionsController}
                                    key={item.action.id}
                                    {...item}
                                    className="mr-1 mb-1"
                                    variant="secondary"
                                    as={ActionItem}
                                />
                            ))}
                        </section>
                    )}
                </ActionsContainer>
            ) : null}

            <ConnectionContainer className={styles.section}>
                <Heading as="h3" styleAs="h2">
                    Changes
                </Heading>

                {error && <ErrorAlert error={error} className="w-100 mb-0" />}
                <ConnectionList className="list-group list-group-flush w-100">
                    {connection?.nodes.map(node => (
                        <GitCommitNode
                            key={node.id}
                            className={classNames('list-group-item', styles.gitCommitNode)}
                            messageSubjectClassName={styles.gitCommitNodeMessageSubject}
                            compact={true}
                            wrapperElement="li"
                            node={node}
                        />
                    ))}
                </ConnectionList>
                {loading && <ConnectionLoading />}
                {connection && (
                    <SummaryContainer centered={true}>
                        <ConnectionSummary
                            centered={true}
                            first={TREE_COMMITS_PER_PAGE}
                            connection={connection}
                            noun={showOlderCommits ? 'commit' : 'commit in the past year'}
                            pluralNoun={showOlderCommits ? 'commits' : 'commits in the past year'}
                            hasNextPage={hasNextPage}
                            emptyElement={null}
                        />
                        {hasNextPage ? (
                            showLinkToCommitsPage ? (
                                <Link to={`${repo.url}/-/commits${filePath ? `/${filePath}` : ''}`}>
                                    Show all commits
                                </Link>
                            ) : (
                                <ShowMoreButton centered={true} onClick={fetchMore} />
                            )
                        ) : null}
                        {!hasNextPage && !showOlderCommits ? showAllCommits : null}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
            {/* ************* */}
        </>
    )
    //     return (
    //         <>
    //             <div>
    //                 {readmeInfo && (
    //                     <ReadmePreviewCard
    //                         readmeHTML={readmeInfo.blob.richHTML}
    //                         readmeURL={readmeInfo.entry.url}
    //                         location={props.location}
    //                     />
    //                 )}
    //             </div>
    //             <section className={classNames('test-tree-entries container mb-3 px-0', styles.section)}>
    //                 <div className="row">
    //                     <div className="col-12 mb-3">
    //                         <FilesCard langStats={langStats} diffStats={diffStats} entries={tree.entries} />
    //                     </div>
    //                 </div>
    //                 <div className="row">
    //                     <div className="col-12 col-lg-6 mb-3">
    //                         <Card className="card">
    //                             <CardHeader>
    //                                 <Link to={`${tree.url}/-/commits`}>
    //                                     Commits
    //                                 </Link>
    //                             </CardHeader>
    //                             {/* TODO(beyang): ultra-compact mode and collapse date timestamps into headers */}
    //                             <FilteredConnection<
    //                                 GitCommitFields,
    //                                 Pick<
    //                                     GitCommitNodeProps,
    //                                     'className' | 'compact' | 'messageSubjectClassName' | 'wrapperElement'
    //                                 >
    //                             >
    //                                 location={props.location}
    //                                 className="foobar"
    //                                 listClassName="list-group list-group-flush"
    //                                 noun="commit in this tree"
    //                                 pluralNoun="commits in this tree"
    //                                 queryConnection={queryCommits}
    //                                 nodeComponent={GitCommitNode}
    //                                 nodeComponentProps={{
    //                                     className: classNames('list-group-item px-2 py-1', styles.gitCommitNode),
    //                                     messageSubjectClassName: styles.gitCommitNodeMessageSubject,
    //                                     compact: true,
    //                                     wrapperElement: 'li',
    //                                 }}
    //                                 updateOnChange={`${repo.name}:${revision}:${filePath}:${String(showOlderCommits)}`}
    //                                 defaultFirst={20}
    //                                 useURLQuery={false}
    //                                 hideSearch={true}
    //                                 emptyElement={emptyElement}
    //                                 totalCountSummaryComponent={TotalCountSummary}
    //                             />
    //                         </Card>
    //                     </div>
    //                     <div className="col-12 col-lg-6">
    //                         <Card className="card">
    //                             <CardHeader>
    //                                 <Link to={`${tree.url}/-/stats/contributors`}>
    //                                     Contributors
    //                                 </Link>
    //                             </CardHeader>
    //                             <Contributors
    //                                 filePath={filePath}
    //                                 tree={tree}
    //                                 repo={repo}
    //                                 commitID={commitID}
    //                                 revision={revision}
    //                                 {...props}
    //                             />
    //                         </Card>
    //                     </div>
    //                 </div>
    //             </section>
    // >>>>>>> dba1cfefd4 (main code changes) */}
}

const CONTRIBUTORS_QUERY = gql`
    query RepositoryContributors($repo: ID!, $first: Int, $revisionRange: String, $afterDate: String, $path: String) {
        node(id: $repo) {
            ... on Repository {
                contributors(first: $first, revisionRange: $revisionRange, afterDate: $afterDate, path: $path) {
                    ...RepositoryContributorConnectionFields
                }
            }
        }
    }

    fragment RepositoryContributorConnectionFields on RepositoryContributorConnection {
        totalCount
        pageInfo {
            hasNextPage
        }
        nodes {
            ...RepositoryContributorNodeFields
        }
    }

    fragment RepositoryContributorNodeFields on RepositoryContributor {
        person {
            name
            displayName
            email
            avatarURL
            user {
                username
                url
                displayName
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

    const { connection, error, loading, hasNextPage, fetchMore } = useConnection<
        RepositoryContributorsResult,
        RepositoryContributorsVariables,
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
            {loading && <ConnectionLoading />}
            <SummaryContainer>
                {connection && (
                    <div className="pl-2">
                        <ConnectionSummary
                            connection={connection}
                            first={BATCH_COUNT}
                            noun="contributor"
                            pluralNoun="contributors"
                            hasNextPage={hasNextPage}
                        />
                    </div>
                )}
                {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
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
    const commit = node.commits.nodes[0] as RepositoryContributorNodeFields['commits']['nodes'][number] | undefined

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
                <UserAvatar inline={true} className="mr-2" user={node.person} />
                <PersonLink userClassName="font-weight-bold" person={node.person} />
            </div>
            <div className={contributorsStyles.commits}>
                <div className={contributorsStyles.commit}>
                    {commit && (
                        <>
                            {/* <Timestamp date={commit.author.date} />:{' '} */}
                            <Tooltip content="Most recent commit by contributor" placement="bottom">
                                <Link to={commit.url} className="repository-contributor-node__commit-subject">
                                    {commit.subject}
                                </Link>
                            </Tooltip>
                        </>
                    )}
                </div>
                <div className={contributorsStyles.count}>
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
            </div>
        </li>
    )
}
