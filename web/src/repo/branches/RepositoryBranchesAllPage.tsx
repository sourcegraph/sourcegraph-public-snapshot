import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../../backend/graphql'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'
import { memoizeObservable } from '../../util/memoize'
import { numberWithCommas } from '../../util/strings'

interface GitRefNodeProps {
    node: GQL.IGitRef
}

export const GitBranchNode: React.SFC<GitRefNodeProps> = ({ node }) => {
    const mostRecentSig =
        node.target.commit &&
        (node.target.commit.committer && node.target.commit.committer.date > node.target.commit.author.date
            ? node.target.commit.committer
            : node.target.commit.author)
    const behindAhead = node.target.commit && node.target.commit.behindAhead
    return (
        <div key={node.id} className="git-branch-node list-group-item">
            <span>
                <Link className="git-ref-tag-2" to={node.url}>
                    <code>{node.displayName}</code>
                </Link>
                {mostRecentSig && (
                    <small className="text-muted pl-2">
                        Updated <Timestamp date={mostRecentSig.date} />{' '}
                        {mostRecentSig.person && <>by {mostRecentSig.person.displayName}</>}
                    </small>
                )}
            </span>
            {behindAhead && (
                <small className="text-muted">
                    {numberWithCommas(behindAhead.behind)} behind, {numberWithCommas(behindAhead.ahead)} ahead
                </small>
            )}
        </div>
    )
}

export const gitBranchFragment = gql`
    fragment GitBranchFields on GitRef {
        id
        displayName
        url
        target {
            commit {
                author {
                    ...SignatureFields
                }
                committer {
                    ...SignatureFields
                }
                behindAhead(revspec: "HEAD") {
                    behind
                    ahead
                }
            }
        }
    }

    fragment SignatureFields on Signature {
        person {
            displayName
        }
        date
    }
`

const fetchGitBranches = memoizeObservable(
    (args: { repo: GQLID; first?: number; query?: string }): Observable<GQL.IGitRefConnection> =>
        queryGraphQL(
            gql`
                query RepositoryGitBranches($repo: ID!, $first: Int, $query: String) {
                    node(id: $repo) {
                        ... on Repository {
                            gitRefs(first: $first, query: $query, type: GIT_BRANCH, orderBy: AUTHORED_OR_COMMITTED_AT) {
                                nodes {
                                    ...GitBranchFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
                ${gitBranchFragment}
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || !(data.node as GQL.IRepository).gitRefs) {
                    throw createAggregateError(errors)
                }
                return (data.node as GQL.IRepository).gitRefs
            })
        ),
    args => `${args.repo}:${args.first}:${args.query}`
)

interface Props {
    repo: GQL.IRepository

    history: H.History
    location: H.Location
}

class FilteredGitRefConnection extends FilteredConnection<GQL.IGitRef> {}

/** A page that shows all of a repository's branches. */
export class RepositoryBranchesAllPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryBranchesAll')
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-branches-page">
                <FilteredGitRefConnection
                    className=""
                    listClassName="list-group list-group-flush"
                    noun="branch"
                    pluralNoun="branches"
                    queryConnection={this.queryBranches}
                    nodeComponent={GitBranchNode}
                    defaultFirst={20}
                    autoFocus={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryBranches = (args: FilteredConnectionQueryArgs) =>
        fetchGitBranches({ ...args, repo: this.props.repo.id })
}
