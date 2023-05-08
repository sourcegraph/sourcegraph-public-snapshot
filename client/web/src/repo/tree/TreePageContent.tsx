import React, { useEffect, useMemo, useState } from 'react'

import { mdiCog, mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'
import { formatISO, subYears } from 'date-fns'
import { escapeRegExp } from 'lodash'
import { Observable } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'

import { RepoMetadata } from '@sourcegraph/branded'
import { encodeURIPathComponent, numberWithCommas, pluralize } from '@sourcegraph/common'
import { dataOrThrowErrors, gql, useQuery } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SearchPatternType, TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Card, CardHeader, Icon, Link, Tooltip, Text, ButtonLink } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    SummaryContainer,
    ConnectionError,
} from '../../components/FilteredConnection/ui'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import {
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
import { GitCommitNodeTableRow } from '../commits/GitCommitNodeTableRow'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'

import { DiffStat, FilesCard, ReadmePreviewCard } from './TreePagePanels'

import styles from './TreePageContent.module.scss'
import contributorsStyles from './TreePageContentContributors.module.scss'
import panelStyles from './TreePagePanels.module.scss'

const COUNT = 20

export interface TreeCommitsResponse {
    ancestors: NonNullable<Extract<TreeCommitsResult['node'], { __typename: 'Repository' }>['commit']>['ancestors']
    externalURLs: Extract<TreeCommitsResult['node'], { __typename: 'Repository' }>['externalURLs']
}

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
        }),
        catchError(() => []) // ignore errors
    )

const ExtraInfoSectionItem: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <div className={styles.extraInfoSectionItem}>{children}</div>
)

const ExtraInfoSectionItemHeader: React.FunctionComponent<
    React.PropsWithChildren<{ title: string; tooltip?: React.ReactNode }>
> = ({ title, tooltip, children }) => (
    <div className="d-flex align-items-center justify-content-between mb-2">
        <div className="d-flex align-items-center">
            <Text className="mr-1 mb-0" weight="bold">
                {title}
            </Text>
            <Tooltip content={tooltip}>
                <Icon
                    svgPath={mdiInformationOutline}
                    aria-label={title}
                    className={classNames('text-muted', styles.extraInfoSectionItemHeaderIcon)}
                />
            </Tooltip>
        </div>
        {children}
    </div>
)

const ExtraInfoSection: React.FC<{
    repo: TreePageRepositoryFields
    className?: string
    viewerCanAdminister?: boolean
}> = ({ repo, className, viewerCanAdminister }) => {
    const [enableRepositoryMetadata] = useFeatureFlag('repository-metadata', false)

    const metadataItems = useMemo(() => repo.metadata.map(({ key, value }) => ({ key, value })) || [], [repo.metadata])

    return (
        <Card className={className}>
            <ExtraInfoSectionItem>
                <ExtraInfoSectionItemHeader title="Description" tooltip="Synced from the code host." />
                {repo.description && <Text>{repo.description}</Text>}
            </ExtraInfoSectionItem>
            {enableRepositoryMetadata && (
                <ExtraInfoSectionItem>
                    <ExtraInfoSectionItemHeader
                        title="Metadata"
                        tooltip={
                            <>
                                Repository metadata allows you to search, filter and navigate between repositories.
                                Administrators can add repository metadata via the web, cli or API. Learn more about{' '}
                                <Link to="/help/admin/repo/metadata" className={styles.linkDark}>
                                    Repository Metadata
                                </Link>
                                .
                            </>
                        }
                    >
                        {viewerCanAdminister && (
                            <Tooltip content="Edit repository metadata">
                                <ButtonLink
                                    to={`/${encodeURIPathComponent(repo.name)}/-/settings/metadata`}
                                    className={classNames('p-0', styles.extraInfoSectionItemHeaderIcon)}
                                >
                                    <Icon
                                        svgPath={mdiCog}
                                        aria-label="Edit repository metadata"
                                        className="text-muted"
                                    />
                                </ButtonLink>
                            </Tooltip>
                        )}
                    </ExtraInfoSectionItemHeader>
                    {metadataItems.length ? (
                        <RepoMetadata items={metadataItems} />
                    ) : (
                        <Text className="text-muted">None</Text>
                    )}
                </ExtraInfoSectionItem>
            )}
        </Card>
    )
}

interface TreePageContentProps extends ExtensionsControllerProps, TelemetryProps, PlatformContextProps {
    filePath: string
    tree: TreeFields
    repo: TreePageRepositoryFields
    commitID: string
    revision: string
    isPackage: boolean
}

export const TreePageContent: React.FunctionComponent<React.PropsWithChildren<TreePageContentProps>> = props => {
    const { filePath, tree, repo, revision, isPackage } = props

    const readmeEntry = useMemo(() => {
        for (const entry of tree.entries) {
            const name = entry.name.toLocaleLowerCase()
            if (!entry.isDirectory && (name === 'readme.md' || name === 'readme' || name === 'readme.txt')) {
                return entry
            }
        }
        return null
    }, [tree.entries])

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

    return (
        <>
            <section className={classNames('container mb-3 px-0', styles.section)}>
                {readmeEntry && (
                    <ReadmePreviewCard
                        entry={readmeEntry}
                        repoName={repo.name}
                        revision={revision}
                        className={styles.files}
                    />
                )}
                <ExtraInfoSection
                    repo={repo}
                    className={classNames(styles.contributors, 'p-3')}
                    viewerCanAdminister={repo.viewerCanAdminister}
                />
            </section>
            <section className={classNames('test-tree-entries container mb-3 px-0', styles.section)}>
                <FilesCard diffStats={diffStats} entries={tree.entries} className={styles.files} filePath={filePath} />

                {!isPackage && (
                    <Card className={styles.commits}>
                        <CardHeader className={panelStyles.cardColHeaderWrapper}>Commits</CardHeader>
                        <Commits {...props} />
                    </Card>
                )}

                {!isPackage && (
                    <Card className={styles.contributors}>
                        <CardHeader className={panelStyles.cardColHeaderWrapper}>Contributors</CardHeader>
                        <Contributors {...props} />
                    </Card>
                )}
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
const Contributors: React.FC<ContributorsProps> = ({ repo, filePath }) => {
    const spec: QuerySpec = {
        revisionRange: '',
        after: '',
        path: filePath,
    }

    const { data, error, loading } = useQuery<
        TreePageRepositoryContributorsResult,
        TreePageRepositoryContributorsVariables
    >(CONTRIBUTORS_QUERY, {
        variables: {
            first: COUNT,
            repo: repo.id,
            revisionRange: spec.revisionRange,
            afterDate: spec.after,
            path: filePath,
        },
    })

    const node = data?.node && data?.node.__typename === 'Repository' ? data.node : null
    const connection = node?.contributors

    return (
        <ConnectionContainer>
            {error && <ConnectionError errors={[error.message]} />}
            {connection && connection.nodes.length > 0 && (
                <ConnectionList
                    className={classNames('test-filtered-contributors-connection', styles.table)}
                    as="table"
                >
                    <tbody>
                        {connection.nodes.map(node => (
                            <RepositoryContributorNode
                                key={node.person.email}
                                node={node}
                                repoName={repo.name}
                                {...spec}
                            />
                        ))}
                    </tbody>
                </ConnectionList>
            )}
            {loading && (
                <div className={contributorsStyles.filteredConnectionLoading}>
                    <ConnectionLoading />
                </div>
            )}
            <SummaryContainer className={styles.tableSummary}>
                {connection && (
                    <>
                        <ConnectionSummary
                            compact={true}
                            connection={connection}
                            first={COUNT}
                            noun="contributor"
                            pluralNoun="contributors"
                            hasNextPage={connection.pageInfo.hasNextPage}
                        />
                        {connection.pageInfo.hasNextPage && (
                            <small>
                                <Link
                                    to={`${repo.url}/-/stats/contributors?${
                                        filePath ? 'path=' + encodeURIComponent(filePath) : ''
                                    }`}
                                >
                                    Show more
                                </Link>
                            </small>
                        )}
                    </>
                )}
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
}
const RepositoryContributorNode: React.FC<RepositoryContributorNodeProps> = ({
    node,
    repoName,
    revisionRange,
    after,
    path,
}) => {
    const query: string = [
        searchQueryForRepoRevision(repoName),
        'type:diff',
        `author:${quoteIfNeeded(node.person.email)}`,
        after ? `after:${quoteIfNeeded(after)}` : '',
        path ? `file:${quoteIfNeeded(escapeRegExp(path))}` : '',
    ]
        .join(' ')
        .replace(/\s+/, ' ')

    return (
        <tr className={classNames('list-group-item', contributorsStyles.repositoryContributorNode)}>
            <td className={contributorsStyles.person}>
                <UserAvatar inline={true} className="mr-2" user={node.person.user ? node.person.user : node.person} />
                <PersonLink person={node.person} />
            </td>
            <td className={contributorsStyles.commits}>
                <Tooltip
                    content={
                        revisionRange?.includes('..')
                            ? 'All commits will be shown (revision end ranges are not yet supported)'
                            : null
                    }
                    placement="left"
                >
                    <Link to={`/search?${buildSearchURLQuery(query, SearchPatternType.standard, false)}`}>
                        {numberWithCommas(node.count)} {pluralize('commit', node.count)}
                    </Link>
                </Tooltip>
            </td>
        </tr>
    )
}

const COMMITS_QUERY = gql`
    query TreeCommits($repo: ID!, $revspec: String!, $first: Int, $filePath: String, $after: String) {
        node(id: $repo) {
            __typename
            ... on Repository {
                externalURLs {
                    url
                    serviceKind
                }
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

interface CommitsProps extends TreePageContentProps {}
const Commits: React.FC<CommitsProps> = ({ repo, revision, filePath, tree }) => {
    const after: string = useMemo(() => formatISO(subYears(Date.now(), 1)), [])
    const { data, error, loading } = useQuery<TreeCommitsResult, TreeCommitsVariables>(COMMITS_QUERY, {
        variables: {
            first: COUNT,
            repo: repo.id,
            revspec: revision || '',
            after,
            filePath,
        },
    })

    const node = data?.node && data?.node.__typename === 'Repository' ? data.node : null
    const connection = node?.commit?.ancestors

    let commitsUrl = tree.url
    if (tree.url.includes('/-/tree')) {
        commitsUrl = commitsUrl.replace('/-/tree', '/-/commits')
    } else {
        commitsUrl = commitsUrl + '/-/commits'
    }

    return (
        <ConnectionContainer>
            {error && <ConnectionError errors={[error.message]} />}
            {connection && connection.nodes.length > 0 && (
                <ConnectionList className={classNames('test-commits-connection', styles.table)} as="table">
                    <tbody>
                        {connection.nodes.map(node => (
                            <GitCommitNodeTableRow
                                key={node.id}
                                node={node}
                                className={styles.gitCommitNode}
                                messageSubjectClassName={styles.gitCommitNodeMessageSubject}
                                compact={true}
                            />
                        ))}
                    </tbody>
                </ConnectionList>
            )}
            {loading && (
                <div className={contributorsStyles.filteredConnectionLoading}>
                    <ConnectionLoading />
                </div>
            )}
            <SummaryContainer className={styles.tableSummary}>
                {connection && (
                    <>
                        <small className="text-muted">
                            <span>
                                {connection.nodes.length > 0 ? (
                                    <>
                                        Showing last {connection.nodes.length}{' '}
                                        {pluralize(
                                            'commit of the past year',
                                            connection.nodes.length,
                                            'commits of the past year'
                                        )}
                                    </>
                                ) : (
                                    <>No commits in the past year</>
                                )}
                            </span>
                        </small>
                        <small>
                            <Link to={commitsUrl}>Show {connection.pageInfo.hasNextPage ? 'more' : 'all'}</Link>
                        </small>
                    </>
                )}
            </SummaryContainer>
        </ConnectionContainer>
    )
}
