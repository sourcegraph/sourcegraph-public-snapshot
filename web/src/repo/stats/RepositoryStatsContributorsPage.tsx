import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { eventLogger } from '../../tracking/eventLogger'
import { PersonLink } from '../../user/PersonLink'
import { UserAvatar } from '../../user/UserAvatar'
import { createAggregateError } from '../../util/errors'
import { memoizeObservable } from '../../util/memoize'
import { pluralize } from '../../util/strings'
import { RepositoryStatsAreaPageProps } from './RepositoryStatsArea'

interface ContributorNodeProps {
    node: GQL.IContributor
}

export const ContributorNode: React.SFC<ContributorNodeProps> = ({ node }) => (
    <div className="contributor-node list-group-item py-2">
        <UserAvatar className="icon-inline mr-1" user={node.person} />
        <PersonLink className="mr-1" userClassName="font-weight-bold" {...node.person} />{' '}
        <span className="badge badge-primary" data-tooltip={`${node.count} ${pluralize('contribution', node.count)}`}>
            {node.count}
        </span>
    </div>
)

const queryContributors = memoizeObservable(
    (args: { repo: GQLID; first?: number; range?: string }): Observable<GQL.IContributorConnection> =>
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

class FilteredContributorsConnection extends FilteredConnection<GQL.IContributor> {}

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
                    queryConnection={this.queryContributors}
                    nodeComponent={ContributorNode}
                    defaultFirst={20}
                    hideFilter={true}
                    autoFocus={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryContributors = (args: { first?: number; range?: string }) =>
        queryContributors({ ...args, repo: this.props.repo.id })
}
