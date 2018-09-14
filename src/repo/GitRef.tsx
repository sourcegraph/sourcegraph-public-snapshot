import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { LinkOrSpan } from '../components/LinkOrSpan'
import { Timestamp } from '../components/time/Timestamp'
import { createAggregateError } from '../util/errors'
import { memoizeObservable } from '../util/memoize'
import { numberWithCommas } from '../util/strings'

interface GitRefNodeProps {
    node: GQL.IGitRef

    /** Link URL; if undefined, node.url is used. */
    url?: string

    /** Whether the top-level element is a link. */
    rootIsLink?: boolean

    children?: React.ReactNode
}

export const GitRefNode: React.SFC<GitRefNodeProps> = ({ node, url, rootIsLink, children }) => {
    const mostRecentSig =
        node.target.commit &&
        (node.target.commit.committer && node.target.commit.committer.date > node.target.commit.author.date
            ? node.target.commit.committer
            : node.target.commit.author)
    const behindAhead = node.target.commit && node.target.commit.behindAhead
    url = url !== undefined ? url : node.url

    return (
        <LinkOrSpan key={node.id} className="git-ref-node list-group-item" to={url}>
            <span>
                {rootIsLink ? (
                    <code className="git-ref-tag-2">{node.displayName}</code>
                ) : (
                    <Link className="git-ref-tag-2" to={url}>
                        <code>{node.displayName}</code>
                    </Link>
                )}
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
            {children}
        </LinkOrSpan>
    )
}

export const gitRefFragment = gql`
    fragment GitRefFields on GitRef {
        id
        displayName
        name
        abbrevName
        url
        target {
            commit {
                author {
                    ...SignatureFields
                }
                committer {
                    ...SignatureFields
                }
                behindAhead(revspec: "HEAD") @include(if: $withBehindAhead) {
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

export const queryGitRefs = memoizeObservable(
    (args: {
        repo: GQL.ID
        first?: number
        query?: string
        type: GQL.GitRefType
        withBehindAhead?: boolean
    }): Observable<GQL.IGitRefConnection> =>
        queryGraphQL(
            gql`
                query RepositoryGitRefs(
                    $repo: ID!
                    $first: Int
                    $query: String
                    $type: GitRefType!
                    $withBehindAhead: Boolean!
                ) {
                    node(id: $repo) {
                        ... on Repository {
                            gitRefs(first: $first, query: $query, type: $type, orderBy: AUTHORED_OR_COMMITTED_AT) {
                                nodes {
                                    ...GitRefFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
                ${gitRefFragment}
            `,
            {
                ...args,
                withBehindAhead:
                    args.withBehindAhead !== undefined ? args.withBehindAhead : args.type === GQL.GitRefType.GIT_BRANCH,
            }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || !(data.node as GQL.IRepository).gitRefs) {
                    throw createAggregateError(errors)
                }
                return (data.node as GQL.IRepository).gitRefs
            })
        ),
    args => `${args.repo}:${args.first}:${args.query}:${args.type}`
)
