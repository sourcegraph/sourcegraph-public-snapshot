import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'
import { RepoHeaderBreadcrumbNavItem } from '../RepoHeaderBreadcrumbNavItem'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'
import { GitCommitNode, GitCommitNodeProps } from './GitCommitNode'

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
        }
        url
        canonicalURL
        externalURLs {
            url
            serviceType
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
        }
        date
    }
`

const fetchGitCommits = (args: {
    repo: GQL.ID
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

interface Props extends RepoHeaderContributionsLifecycleProps {
    repo: GQL.IRepository
    rev?: string
    commitID: string

    history: H.History
    location: H.Location
}

/** A page that shows a repository's commits at the current revision. */
export class RepositoryCommitsPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryCommits')
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-commits-page">
                <PageTitle title="Commits" />
                <RepoHeaderContributionPortal
                    position="nav"
                    element={<RepoHeaderBreadcrumbNavItem key="commits">Commits</RepoHeaderBreadcrumbNavItem>}
                    repoHeaderContributionsLifecycleProps={this.props.repoHeaderContributionsLifecycleProps}
                />
                <FilteredConnection<GQL.IGitCommit, Pick<GitCommitNodeProps, 'repoName' | 'className' | 'compact'>>
                    className="repository-commits-page__content"
                    listClassName="list-group list-group-flush"
                    noun="commit"
                    pluralNoun="commits"
                    queryConnection={this.queryCommits}
                    nodeComponent={GitCommitNode}
                    nodeComponentProps={{ repoName: this.props.repo.name, className: 'list-group-item' }}
                    defaultFirst={20}
                    autoFocus={true}
                    history={this.props.history}
                    hideSearch={true}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryCommits = (args: FilteredConnectionQueryArgs) =>
        fetchGitCommits({ ...args, repo: this.props.repo.id, revspec: this.props.commitID })
}
