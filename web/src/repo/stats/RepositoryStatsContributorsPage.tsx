import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { PersonLink } from '../../user/PersonLink'
import { UserAvatar } from '../../user/UserAvatar'
import { createAggregateError } from '../../util/errors'
import { memoizeObservable } from '../../util/memoize'
import { pluralize } from '../../util/strings'
import { RepositoryStatsAreaPageProps } from './RepositoryStatsArea'

interface ContributorNodeProps {
    node: GQL.IRepositoryContributor
}

export const RepositoryContributorNode: React.SFC<ContributorNodeProps> = ({ node }) => {
    const commit = node.commits.nodes[0]
    return (
        <div className="repository-contributor-node list-group-item py-2">
            <div className="repository-contributor-node__count">
                <span
                    className="badge badge-primary"
                    data-tooltip={`${node.count} ${pluralize('contribution', node.count)}`}
                >
                    {node.count}
                </span>
            </div>
            <div className="repository-contributor-node__person">
                <UserAvatar className="icon-inline mr-2" user={node.person} />
                <PersonLink userClassName="font-weight-bold" {...node.person} />
            </div>
            <div className="repository-contributor-node__commit">
                <Timestamp date={commit.author.date} />:{' '}
                <Link to={commit.url} className="repository-contributor-node__commit-subject">
                    {commit.subject}
                </Link>
            </div>
        </div>
    )
}

const queryRepositoryContributors = memoizeObservable(
    (args: { repo: GQLID; first?: number; range?: string }): Observable<GQL.IRepositoryContributorConnection> =>
        queryGraphQL(
            gql`
                query RepositoryContributors($repo: ID!, $first: Int, $range: String) {
                    node(id: $repo) {
                        ... on Repository {
                            contributors(first: $first, range: $range) {
                                nodes {
                                    person {
                                        name
                                        displayName
                                        email
                                        avatarURL
                                        user {
                                            username
                                            url
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
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || !(data.node as GQL.IRepository).contributors || errors) {
                    throw createAggregateError(errors)
                }
                return (data.node as GQL.IRepository).contributors
            })
        ),
    args => `${args.repo}:${args.first}:${args.range}`
)

interface Props extends RepositoryStatsAreaPageProps, RouteComponentProps<{}> {}

class FilteredContributorsConnection extends FilteredConnection<GQL.IRepositoryContributor> {}

/** A page that shows a repository's contributors. */
export class RepositoryStatsContributorsPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryStatsContributors')
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-stats-page">
                <FilteredContributorsConnection
                    className=""
                    listClassName="list-group list-group-flush"
                    noun="contributor"
                    pluralNoun="contributors"
                    queryConnection={this.queryRepositoryContributors}
                    nodeComponent={RepositoryContributorNode}
                    defaultFirst={20}
                    hideFilter={true}
                    autoFocus={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryRepositoryContributors = (args: { first?: number; range?: string }) =>
        queryRepositoryContributors({ ...args, repo: this.props.repo.id })
}
