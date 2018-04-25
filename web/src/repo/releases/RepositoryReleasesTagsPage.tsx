import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'
import { memoizeObservable } from '../../util/memoize'

interface GitTagNodeProps {
    node: GQL.IGitRef
}

export const GitTagNode: React.SFC<GitTagNodeProps> = ({ node }) => {
    const mostRecentSig =
        node.target.commit &&
        (node.target.commit.committer && node.target.commit.committer.date > node.target.commit.author.date
            ? node.target.commit.committer
            : node.target.commit.author)
    return (
        <div key={node.id} className="git-tag-node list-group-item">
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
        </div>
    )
}

const gitTagFragment = gql`
    fragment GitTagFields on GitRef {
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

const fetchGitTags = memoizeObservable(
    (args: { repo: GQL.ID; first?: number; query?: string }): Observable<GQL.IGitRefConnection> =>
        queryGraphQL(
            gql`
                query RepositoryGitTags($repo: ID!, $first: Int, $query: String) {
                    node(id: $repo) {
                        ... on Repository {
                            gitRefs(first: $first, query: $query, type: GIT_TAG, orderBy: AUTHORED_OR_COMMITTED_AT) {
                                nodes {
                                    ...GitTagFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
                ${gitTagFragment}
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

/** A page that shows all of a repository's tags. */
export class RepositoryReleasesTagsPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryReleasesTags')
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-releases-page">
                <PageTitle title="Tags" />
                <FilteredGitRefConnection
                    className=""
                    listClassName="list-group list-group-flush"
                    noun="tag"
                    pluralNoun="tags"
                    queryConnection={this.queryTags}
                    nodeComponent={GitTagNode}
                    defaultFirst={20}
                    autoFocus={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryTags = (args: FilteredConnectionQueryArgs) => fetchGitTags({ ...args, repo: this.props.repo.id })
}
