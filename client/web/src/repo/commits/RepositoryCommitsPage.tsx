import React, { useEffect, useCallback, useMemo } from 'react'

import * as H from 'history'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import { RevisionSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'

import { queryGraphQL } from '../../backend/graphql'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { GitCommitFields, RepositoryFields, Scalars } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { externalLinkFieldsFragment } from '../backend'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'

import { GitCommitNode, GitCommitNodeProps } from './GitCommitNode'

import styles from './RepositoryCommitsPage.module.scss'

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
}): Observable<GQL.IGitCommitConnection> =>
    queryGraphQL(
        gql`
            query RepositoryGitCommits($repo: ID!, $revspec: String!, $first: Int, $query: String) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $revspec) {
                            ancestors(first: $first, query: $query) {
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
            const repo = data.node as GQL.IRepository
            if (!repo.commit || !repo.commit.ancestors) {
                throw createAggregateError(errors)
            }
            return repo.commit.ancestors
        })
    )

interface Props
    extends RepoHeaderContributionsLifecycleProps,
        Partial<RevisionSpec>,
        ResolvedRevisionSpec,
        BreadcrumbSetters {
    repo: RepositoryFields

    history: H.History
    location: H.Location
}

/** A page that shows a repository's commits at the current revision. */
export const RepositoryCommitsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    useBreadcrumb,
    ...props
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('RepositoryCommits')
    }, [])

    useBreadcrumb(useMemo(() => ({ key: 'commits', element: <>Commits</> }), []))

    const queryCommits = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<GQL.IGitCommitConnection> =>
            fetchGitCommits({ ...args, repo: props.repo.id, revspec: props.commitID }),
        [props.repo.id, props.commitID]
    )

    return (
        <div className={styles.repositoryCommitsPage} data-testid="commits-page">
            <PageTitle title="Commits" />
            <FilteredConnection<GitCommitFields, Pick<GitCommitNodeProps, 'className' | 'compact' | 'wrapperElement'>>
                className={styles.content}
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
    )
}
