import { useEffect, useMemo, type FC } from 'react'

import { capitalize } from 'lodash'
import { useLocation } from 'react-router-dom'

import { basename, pluralize } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import type { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Code, ErrorAlert, Heading } from '@sourcegraph/wildcard'

import type { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { useUrlSearchParamsForConnectionState } from '../../components/FilteredConnection/hooks/connectionState'
import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'
import { PageTitle } from '../../components/PageTitle'
import {
    RepositoryType,
    type GitCommitFields,
    type RepositoryFields,
    type RepositoryGitCommitsResult,
    type RepositoryGitCommitsVariables,
} from '../../graphql-operations'
import { parseBrowserRepoURL } from '../../util/url'
import { externalLinkFieldsFragment } from '../backend'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'
import { getRefType, isPerforceDepotSource } from '../utils'

import { GitCommitNode } from './GitCommitNode'

import styles from './RepositoryCommitsPage.module.scss'

export const gitCommitFragment = gql`
    fragment GitCommitFields on GitCommit {
        id
        oid
        abbreviatedOID
        perforceChangelist {
            ...PerforceChangelistFieldsWithoutCommit
        }
        message
        subject
        body
        author {
            ...SignatureFields
        }
        committer {
            ...SignatureFields
        }
        parents {
            oid
            abbreviatedOID
            perforceChangelist {
                ...PerforceChangelistFieldsWithoutCommit
            }
            url
        }
        url
        canonicalURL
        externalURLs {
            ...ExternalLinkFields
        }
        tree(path: "") {
            canonicalURL
        }
    }

    fragment PerforceChangelistFieldsWithoutCommit on PerforceChangelist {
        cid
        canonicalURL
    }

    fragment SignatureFields on Signature {
        person {
            avatarURL
            name
            email
            displayName
            user {
                id
                username
                url
                displayName
            }
        }
        date
    }
    ${externalLinkFieldsFragment}
`

export const REPOSITORY_GIT_COMMITS_QUERY = gql`
    query RepositoryGitCommits($repo: ID!, $revspec: String!, $first: Int, $afterCursor: String, $filePath: String) {
        node(id: $repo) {
            ... on Repository {
                sourceType
                externalURLs {
                    url
                    serviceKind
                }
                commit(rev: $revspec) {
                    ancestors(first: $first, path: $filePath, afterCursor: $afterCursor) {
                        nodes {
                            ...GitCommitFields
                        }
                        pageInfo {
                            hasNextPage
                            endCursor
                        }
                    }
                }
            }
        }
    }
    ${gitCommitFragment}
`

export interface RepositoryCommitsPageProps extends RevisionSpec, BreadcrumbSetters, TelemetryProps, TelemetryV2Props {
    repo: RepositoryFields
}

// A page that shows a repository's commits at the current revision.
export const RepositoryCommitsPage: FC<RepositoryCommitsPageProps> = props => {
    const { useBreadcrumb, repo } = props
    const location = useLocation()
    const { filePath = '' } = parseBrowserRepoURL(location.pathname)

    let sourceType = RepositoryType.GIT_REPOSITORY

    const connectionState = useUrlSearchParamsForConnectionState([])
    const { connection, error, loading, hasNextPage, fetchMore } = useShowMorePagination<
        RepositoryGitCommitsResult,
        RepositoryGitCommitsVariables,
        GitCommitFields
    >({
        query: REPOSITORY_GIT_COMMITS_QUERY,
        variables: {
            repo: repo.id,
            revspec: props.revision,
            filePath: filePath ?? null,
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

            sourceType = node.sourceType

            return node?.commit?.ancestors
        },
        options: {
            fetchPolicy: 'cache-first',
            useAlternateAfterCursor: true,
            errorPolicy: 'all',
        },
        state: connectionState,
    })

    const getPageTitle = (): string => {
        const repoString = displayRepoName(repo.name)
        const refType = capitalize(pluralize(getRefType(sourceType), 0))
        if (filePath) {
            return `${refType} - ${basename(filePath)} - ${repoString}`
        }
        return `${refType} - ${repoString}`
    }

    useEffect(() => {
        EVENT_LOGGER.logPageView('RepositoryCommits')
        props.telemetryRecorder.recordEvent('repo.commits', 'view')
    }, [props.telemetryRecorder])

    useBreadcrumb(
        useMemo(() => {
            if (!filePath || !repo) {
                return
            }
            return {
                key: 'treePath',
                className: 'flex-shrink-past-contents',
                element: (
                    <FilePathBreadcrumbs
                        key="path"
                        repoName={repo.name}
                        revision={props.revision}
                        filePath={filePath}
                        isDir={true}
                        telemetryService={props.telemetryService}
                        telemetryRecorder={props.telemetryRecorder}
                    />
                ),
            }
        }, [filePath, repo, props.revision, props.telemetryService, props.telemetryRecorder])
    )
    // We need to resolve the Commits breadcrumb at the same time as the
    // filePath, so that the order is correct (otherwise Commits will show
    // before the path)
    useBreadcrumb(
        useMemo(() => {
            if (!repo) {
                return
            }

            const refType = getRefType(sourceType)

            return {
                key: refType,
                element: <>{capitalize(pluralize(refType, 0))}</>,
            }
        }, [repo, sourceType])
    )

    return (
        <div className={styles.repositoryCommitsPage} data-testid="commits-page">
            <PageTitle title={getPageTitle()} />

            <div className={styles.content}>
                <ConnectionContainer>
                    <Heading as="h2" styleAs="h1">
                        {filePath ? (
                            <>
                                View {pluralize(getRefType(sourceType), 0)} inside <Code>{basename(filePath)}</Code>
                            </>
                        ) : (
                            <>
                                View {pluralize(getRefType(sourceType), 0)} from this{' '}
                                {isPerforceDepotSource(sourceType) ? 'depot' : 'repository'}
                            </>
                        )}
                    </Heading>

                    <Heading as="h3" styleAs="h2">
                        Changes
                    </Heading>

                    {error && <ErrorAlert error={error} className="w-100 mb-0" />}
                    <ConnectionList className="list-group list-group-flush w-100">
                        {connection?.nodes.map(node => (
                            <GitCommitNode
                                key={node.id}
                                className="list-group-item"
                                wrapperElement="li"
                                node={node}
                                telemetryRecorder={props.telemetryRecorder}
                            />
                        ))}
                    </ConnectionList>
                    {loading && <ConnectionLoading />}
                    {connection && (
                        <SummaryContainer centered={true}>
                            <ConnectionSummary
                                centered={true}
                                connection={connection}
                                noun={getRefType(sourceType)}
                                pluralNoun={pluralize(getRefType(sourceType), 0)}
                                hasNextPage={hasNextPage}
                                emptyElement={null}
                            />
                            {hasNextPage ? <ShowMoreButton centered={true} onClick={fetchMore} /> : null}
                        </SummaryContainer>
                    )}
                </ConnectionContainer>
            </div>
        </div>
    )
}
