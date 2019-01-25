import formatDistanceStrict from 'date-fns/formatDistanceStrict'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { ChatIcon } from '../../../../shared/src/components/icons'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'

interface Props {}

const LOADING: 'loading' = 'loading'

interface State {
    /** The threads, loading, or an error. */
    threadsOrError: typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike
}

/**
 * An explore section that shows recent discussion threads.
 */

export class DiscussionsExploreSection extends React.PureComponent<Props, State> {
    private static QUERY_DISCUSSIONS_ARG_FIRST = 4

    public state: State = { threadsOrError: LOADING }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            queryDiscussionThreads({ first: DiscussionsExploreSection.QUERY_DISCUSSIONS_ARG_FIRST })
                .pipe(catchError(err => [asError(err)]))
                .subscribe(threadsOrError => this.setState({ threadsOrError }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const threadsOrError: (typeof LOADING | GQL.IDiscussionThread)[] | ErrorLike =
            this.state.threadsOrError === LOADING
                ? Array(DiscussionsExploreSection.QUERY_DISCUSSIONS_ARG_FIRST).fill(LOADING)
                : isErrorLike(this.state.threadsOrError)
                ? this.state.threadsOrError
                : this.state.threadsOrError.nodes

        const itemClass = 'py-2 border-white'
        return (
            <div className="discussions-explore-section">
                <h2>Recent discussions</h2>
                {isErrorLike(threadsOrError) ? (
                    <div className="alert alert-danger">Error: {threadsOrError.message}</div>
                ) : threadsOrError.length === 0 ? (
                    <p>
                        No discussion threads. Start a discussion by clicking the <ChatIcon className="icon-inline" />{' '}
                        icon next to a line number in a file.
                    </p>
                ) : (
                    <>
                        <div className="discussions-explore-section__list">
                            {threadsOrError.map((thread /* or loading */, i) =>
                                thread === LOADING ? (
                                    <div key={i} className={`${itemClass} list-group-item`}>
                                        <h3 className="text-muted mb-0">⋯</h3>&nbsp;
                                    </div>
                                ) : (
                                    <LinkOrSpan
                                        key={i}
                                        className={`d-flex align-items-center justify-content-between discussions-explore-section__list__item`}
                                        to={thread.inlineURL}
                                    >
                                        <div>
                                            <h3 className="mb-0 text-truncate discussions-explore-section__list__title">
                                                {thread.title} <small>#{thread.id}</small>
                                            </h3>
                                            Created{' '}
                                            {formatDistanceStrict(thread.updatedAt, Date.now(), {
                                                addSuffix: true,
                                            })}
                                            {uniqueAuthors(thread.comments.nodes).map(user => (
                                                <span key={user.username} className="mr-1">
                                                    {' '}
                                                    by {user.username}
                                                </span>
                                            ))}{' '}
                                            <br />
                                            {thread.target.repository.name}
                                        </div>
                                        <div className="h4 mb-0">{thread.comments.totalCount}</div>
                                    </LinkOrSpan>
                                )
                            )}
                        </div>
                        <div className="text-right mt-3">
                            <Link to="/discussions">
                                View all discussions
                                <ChevronRightIcon className="icon-inline" />
                            </Link>
                        </div>
                    </>
                )}
            </div>
        )
    }
}

function uniqueAuthors(comments: GQL.IDiscussionComment[]): GQL.IUser[] {
    const seen = new Set<string>()
    const users: GQL.IUser[] = []
    for (const comment of comments) {
        const key = comment.author.username
        if (!seen.has(key)) {
            users.push(comment.author)
            seen.add(key)
        }
    }
    return users
}

function queryDiscussionThreads(
    args: Pick<GQL.IDiscussionThreadsOnQueryArguments, 'first'>
): Observable<GQL.IDiscussionThreadConnection> {
    return queryGraphQL(
        gql`
            query DiscussionThreads($first: Int) {
                discussionThreads(first: $first) {
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                    nodes {
                        ...DiscussionThreadFields
                        comments(first: 10) {
                            nodes {
                                author {
                                    username
                                }
                            }
                            totalCount
                        }
                    }
                }
            }

            fragment DiscussionThreadFields on DiscussionThread {
                id
                author {
                    ...UserFields
                }
                title
                target {
                    __typename
                    ... on DiscussionThreadTargetRepo {
                        repository {
                            name
                        }
                    }
                }
                inlineURL
                createdAt
                updatedAt
                archivedAt
            }

            fragment UserFields on User {
                displayName
                username
                avatarURL
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.discussionThreads || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.discussionThreads
        })
    )
}
