import * as React from 'react'
import { useCallback } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import FileIcon from 'mdi-react/FileIcon'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createInvalidGraphQLQueryResponseError, dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { RevisionSpec, FileSpec } from '@sourcegraph/shared/src/util/url'
import { Link, Icon } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../backend/graphql'
import { FilteredConnection } from '../components/FilteredConnection'
import {
    CommitAncestorsConnectionFields,
    FetchCommitsResult,
    FetchCommitsVariables,
    GitCommitFields,
    Scalars,
} from '../graphql-operations'
import { replaceRevisionInURL } from '../util/url'

import { GitCommitNode } from './commits/GitCommitNode'
import { gitCommitFragment } from './commits/RepositoryCommitsPage'

import styles from './RepoRevisionSidebarCommits.module.scss'

interface CommitNodeProps {
    node: GitCommitFields
    location: H.Location
    preferAbsoluteTimestamps: boolean
}

const CommitNode: React.FunctionComponent<React.PropsWithChildren<CommitNodeProps>> = ({
    node,
    location,
    preferAbsoluteTimestamps,
}) => (
    <li className={classNames(styles.commitContainer, 'list-group-item p-0')}>
        <GitCommitNode
            className={styles.commitNode}
            compact={true}
            node={node}
            hideExpandCommitMessageBody={true}
            preferAbsoluteTimestamps={preferAbsoluteTimestamps}
            afterElement={
                <Link
                    to={replaceRevisionInURL(location.pathname + location.search + location.hash, node.oid)}
                    className={classNames(styles.fileIcon, 'ml-2')}
                    title="View current file at this commit"
                >
                    <Icon role="img" as={FileIcon} aria-hidden={true} />
                </Link>
            }
        />
    </li>
)

interface Props extends Partial<RevisionSpec>, FileSpec {
    repoID: Scalars['ID']
    history: H.History
    location: H.Location
    preferAbsoluteTimestamps: boolean
}

export const RepoRevisionSidebarCommits: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const queryCommits = useCallback(
        (args: { query?: string }): Observable<CommitAncestorsConnectionFields> =>
            fetchCommits(props.repoID, props.revision || '', { ...args, currentPath: props.filePath || '' }),
        [props.repoID, props.revision, props.filePath]
    )

    return (
        <FilteredConnection<
            GitCommitFields,
            Pick<CommitNodeProps, 'location' | 'preferAbsoluteTimestamps'>,
            CommitAncestorsConnectionFields
        >
            className="list-group list-group-flush"
            listClassName={styles.list}
            summaryClassName={styles.summary}
            loaderClassName={styles.loader}
            compact={true}
            noun="commit"
            pluralNoun="commits"
            queryConnection={queryCommits}
            nodeComponent={CommitNode}
            nodeComponentProps={{ location: props.location, preferAbsoluteTimestamps: props.preferAbsoluteTimestamps }}
            defaultFirst={100}
            hideSearch={true}
            useURLQuery={false}
            history={props.history}
            location={props.location}
        />
    )
}

function fetchCommits(
    repo: Scalars['ID'],
    revision: string,
    args: { first?: number; currentPath?: string; query?: string }
): Observable<CommitAncestorsConnectionFields> {
    return requestGraphQL<FetchCommitsResult, FetchCommitsVariables>(
        gql`
            query FetchCommits($repo: ID!, $revision: String!, $first: Int, $currentPath: String, $query: String) {
                node(id: $repo) {
                    __typename
                    ... on Repository {
                        commit(rev: $revision) {
                            ancestors(first: $first, query: $query, path: $currentPath) {
                                ...CommitAncestorsConnectionFields
                            }
                        }
                    }
                }
            }

            ${gitCommitFragment}

            fragment CommitAncestorsConnectionFields on GitCommitConnection {
                nodes {
                    ...GitCommitFields
                }
                pageInfo {
                    hasNextPage
                }
            }
        `,
        {
            currentPath: args.currentPath ?? null,
            first: args.first ?? null,
            query: args.query ?? null,
            repo,
            revision,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node || data.node.__typename !== 'Repository' || !data.node.commit) {
                throw createInvalidGraphQLQueryResponseError('FetchCommits')
            }
            return data.node.commit.ancestors
        })
    )
}
