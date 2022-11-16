import React, { useEffect, useCallback, useMemo } from 'react'

import * as H from 'history'
import { RouteComponentProps } from 'react-router'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Code, Heading, LoadingSpinner } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { GitCommitFields, RepositoryFields, RepositoryGitCommitsResult, Scalars } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { basename } from '../../util/path'
import { externalLinkFieldsFragment } from '../backend'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'

import { GitCommitNode, GitCommitNodeProps } from './GitCommitNode'

import styles from './RepositoryCommitsPage.module.scss'

type RepositoryGitCommitsRepository = Extract<RepositoryGitCommitsResult['node'], { __typename?: 'Repository' }>

export const gitCommitFragment = gql`
    fragment GitCommitFields on GitCommit {
        id
        oid
        abbreviatedOID
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

const fetchGitCommits = (args: {
    repo: Scalars['ID']
    revspec: string
    first?: number
    query?: string
    filePath?: string
}): Observable<NonNullable<RepositoryGitCommitsRepository['commit']>['ancestors']> =>
    queryGraphQL(
        gql`
            query RepositoryGitCommits($repo: ID!, $revspec: String!, $first: Int, $query: String, $filePath: String) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $revspec) {
                            ancestors(first: $first, query: $query, path: $filePath) {
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
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node) {
                throw createAggregateError(errors)
            }
            const repo = data.node as RepositoryGitCommitsRepository
            if (!repo.commit || !repo.commit.ancestors) {
                throw createAggregateError(errors)
            }
            return repo.commit.ancestors
        })
    )

export interface RepositoryCommitsPageProps
    extends RevisionSpec,
        BreadcrumbSetters,
        RouteComponentProps<{
            filePath?: string | undefined
        }>,
        TelemetryProps {
    repo: RepositoryFields | undefined

    history: H.History
    location: H.Location
}

// A page that shows a repository's commits at the current revision.
export const RepositoryCommitsPage: React.FunctionComponent<React.PropsWithChildren<RepositoryCommitsPageProps>> = ({
    useBreadcrumb,
    ...props
}) => {
    const repo = props.repo
    const filePath = props.match.params.filePath

    useEffect(() => {
        eventLogger.logPageView('RepositoryCommits')
    }, [])

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
                    />
                ),
            }
        }, [filePath, repo, props.revision, props.telemetryService])
    )
    // We need to resolve the Commits breadcrumb at the same time as the
    // filePath, so that the order is correct (otherwise Commits will show
    // before the path)
    useBreadcrumb(
        useMemo(() => {
            if (!repo) {
                return
            }
            return { key: 'commits', element: <>Commits</> }
        }, [repo])
    )

    const queryCommits = useCallback(
        (
            args: FilteredConnectionQueryArguments
        ): Observable<NonNullable<RepositoryGitCommitsRepository['commit']>['ancestors']> => {
            if (!repo?.id) {
                return of()
            }

            return fetchGitCommits({ ...args, repo: repo?.id, revspec: props.revision, filePath })
        },
        [repo?.id, props.revision, filePath]
    )

    if (!repo) {
        return <LoadingSpinner />
    }

    const getPageTitle = (): string => {
        const repoString = displayRepoName(repo.name)
        if (filePath) {
            return `Commits - ${basename(filePath)} - ${repoString}`
        }
        return `Commits - ${repoString}`
    }

    return (
        <div className={styles.repositoryCommitsPage} data-testid="commits-page">
            <PageTitle title={getPageTitle()} />

            <div className={styles.content}>
                <Heading as="h2" styleAs="h1">
                    {filePath ? (
                        <>
                            View commits inside <Code>{basename(filePath)}</Code>
                        </>
                    ) : (
                        <>View commits from this repository</>
                    )}
                </Heading>
                <FilteredConnection<
                    GitCommitFields,
                    Pick<GitCommitNodeProps, 'className' | 'compact' | 'wrapperElement'>
                >
                    listClassName="list-group list-group-flush"
                    noun="commit"
                    pluralNoun="commits"
                    queryConnection={queryCommits}
                    nodeComponent={GitCommitNode}
                    nodeComponentProps={{ className: 'list-group-item', wrapperElement: 'li' }}
                    defaultFirst={20}
                    autoFocus={true}
                    history={props.history}
                    hideSearch={true}
                    location={props.location}
                />
            </div>
        </div>
    )
}
