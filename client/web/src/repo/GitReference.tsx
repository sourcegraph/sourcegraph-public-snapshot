import classNames from 'classnames'
import * as React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { memoizeObservable } from '@sourcegraph/shared/src/util/memoizeObservable'
import { numberWithCommas } from '@sourcegraph/shared/src/util/strings'
import { Badge } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../backend/graphql'
import { Timestamp } from '../components/time/Timestamp'
import {
    GitRefConnectionFields,
    GitRefFields,
    GitRefType,
    RepositoryGitRefsResult,
    RepositoryGitRefsVariables,
    Scalars,
} from '../graphql-operations'

import styles from './GitReference.module.scss'

export interface GitReferenceNodeProps {
    node: GitRefFields

    /** Link URL; if undefined, node.url is used. */
    url?: string

    /** Whether any ancestor element higher up in the tree is an `<a>` element. */
    ancestorIsLink?: boolean

    children?: React.ReactNode

    className?: string

    icon?: React.ComponentType<{ className?: string }>

    onClick?: React.MouseEventHandler<HTMLAnchorElement>
}

export const GitReferenceNode: React.FunctionComponent<GitReferenceNodeProps> = ({
    node,
    url,
    ancestorIsLink,
    children,
    className,
    onClick,
    icon: Icon,
}) => {
    const mostRecentSig =
        node.target.commit &&
        (node.target.commit.committer && node.target.commit.committer.date > node.target.commit.author.date
            ? node.target.commit.committer
            : node.target.commit.author)
    const behindAhead = node.target.commit?.behindAhead
    url = url !== undefined ? url : node.url

    return (
        <LinkOrSpan
            key={node.id}
            className={classNames('list-group-item', styles.gitRefNode, className)}
            to={!ancestorIsLink ? url : undefined}
            onClick={onClick}
            data-testid="git-ref-node"
        >
            <span className="d-flex align-items-center">
                {Icon && <Icon className="icon-inline mr-1" />}
                <Badge as="code">{node.displayName}</Badge>
                {mostRecentSig && (
                    <small className="pl-2">
                        Updated <Timestamp date={mostRecentSig.date} />{' '}
                        {mostRecentSig.person && <>by {mostRecentSig.person.displayName}</>}
                    </small>
                )}
            </span>
            {behindAhead && (
                <small>
                    {numberWithCommas(behindAhead.behind)} behind, {numberWithCommas(behindAhead.ahead)} ahead
                </small>
            )}
            {children}
        </LinkOrSpan>
    )
}

export const gitReferenceFragments = gql`
    fragment GitRefFields on GitRef {
        id
        displayName
        name
        abbrevName
        url
        target {
            commit {
                author {
                    ...SignatureFieldsForReferences
                }
                committer {
                    ...SignatureFieldsForReferences
                }
                behindAhead(revspec: "HEAD") @include(if: $withBehindAhead) {
                    behind
                    ahead
                }
            }
        }
    }

    fragment SignatureFieldsForReferences on Signature {
        __typename
        person {
            displayName
            user {
                username
            }
        }
        date
    }
`

export const REPOSITORY_GIT_REFS = gql`
    query RepositoryGitRefs($repo: ID!, $first: Int, $query: String, $type: GitRefType!, $withBehindAhead: Boolean!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                gitRefs(first: $first, query: $query, type: $type, orderBy: AUTHORED_OR_COMMITTED_AT) {
                    __typename
                    ...GitRefConnectionFields
                }
            }
        }
    }

    fragment GitRefConnectionFields on GitRefConnection {
        nodes {
            __typename
            ...GitRefFields
        }
        totalCount
        pageInfo {
            hasNextPage
        }
    }

    ${gitReferenceFragments}
`

export const queryGitReferences = memoizeObservable(
    (args: {
        repo: Scalars['ID']
        first?: number
        query?: string
        type: GitRefType
        withBehindAhead?: boolean
    }): Observable<GitRefConnectionFields> =>
        requestGraphQL<RepositoryGitRefsResult, RepositoryGitRefsVariables>(REPOSITORY_GIT_REFS, {
            query: args.query ?? null,
            first: args.first ?? null,
            repo: args.repo,
            type: args.type,
            withBehindAhead:
                args.withBehindAhead !== undefined ? args.withBehindAhead : args.type === GitRefType.GIT_BRANCH,
        }).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || data.node.__typename !== 'Repository' || !data.node.gitRefs) {
                    throw createAggregateError(errors)
                }
                return data.node.gitRefs
            })
        ),
    args => `${args.repo}:${String(args.first)}:${String(args.query)}:${args.type}`
)
