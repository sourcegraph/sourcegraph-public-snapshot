import * as React from 'react'

import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import prettyBytes from 'pretty-bytes'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { map, switchMap, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { createAggregateError, pluralize } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Container, PageHeader, LoadingSpinner, Link, Alert, Icon, Typography } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { Scalars, SettingsAreaRepositoryFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import styles from './RepoSettingsIndexPage.module.scss'

/**
 * Fetches a repository's text search index information.
 */
function fetchRepositoryTextSearchIndex(id: Scalars['ID']): Observable<GQL.IRepositoryTextSearchIndex | null> {
    return queryGraphQL(
        gql`
            query RepositoryTextSearchIndex($id: ID!) {
                node(id: $id) {
                    ... on Repository {
                        textSearchIndex {
                            status {
                                updatedAt
                                contentByteSize
                                contentFilesCount
                                indexByteSize
                                indexShardsCount
                                newLinesCount
                                defaultBranchNewLinesCount
                                otherBranchesNewLinesCount
                            }
                            refs {
                                ref {
                                    displayName
                                    url
                                }
                                indexed
                                current
                                indexedCommit {
                                    oid
                                    abbreviatedOID
                                    commit {
                                        url
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node || errors) {
                throw createAggregateError(errors)
            }
            return (data.node as GQL.IRepository).textSearchIndex
        })
    )
}

const TextSearchIndexedReference: React.FunctionComponent<
    React.PropsWithChildren<{
        repo: SettingsAreaRepositoryFields
        indexedRef: GQL.IRepositoryTextSearchIndexedRef
    }>
> = ({ repo, indexedRef }) => {
    const isCurrent = indexedRef.indexed && indexedRef.current

    return (
        <li className={styles.ref}>
            <Icon
                className={classNames(styles.refIcon, isCurrent && styles.refIconCurrent)}
                as={isCurrent ? CheckCircleIcon : LoadingSpinner}
            />
            <LinkOrSpan to={indexedRef.ref.url}>
                <strong>
                    <code>{indexedRef.ref.displayName}</code>
                </strong>
            </LinkOrSpan>{' '}
            {indexedRef.indexed ? (
                <span>
                    &nbsp;&mdash; indexed at{' '}
                    <code>
                        <LinkOrSpan
                            to={indexedRef.indexedCommit?.commit ? indexedRef.indexedCommit.commit.url : repo.url}
                        >
                            {indexedRef.indexedCommit!.abbreviatedOID}
                        </LinkOrSpan>
                    </code>{' '}
                    {indexedRef.current ? '(up to date)' : '(index update in progress)'}
                </span>
            ) : (
                <span>&nbsp;&mdash; initial indexing in progress</span>
            )}
        </li>
    )
}

interface Props extends RouteComponentProps<{}> {
    repo: SettingsAreaRepositoryFields
}

interface State {
    textSearchIndex?: GQL.IRepositoryTextSearchIndex | null
    loading: boolean
    error?: Error
}

function prettyBytesBigint(bytes: bigint): string {
    let unit = 0
    const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    const threshold = BigInt(1000)

    while (bytes >= threshold) {
        bytes /= threshold
        unit += 1
    }

    return bytes.toString() + ' ' + units[unit]
}

/**
 * The repository settings index page.
 */
export class RepoSettingsIndexPage extends React.PureComponent<Props, State> {
    public state: State = { loading: true }

    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettingsIndex')

        this.subscriptions.add(
            this.updates
                .pipe(
                    tap(() => this.setState({ loading: true })),
                    switchMap(() => fetchRepositoryTextSearchIndex(this.props.repo.id))
                )
                .subscribe(
                    textSearchIndex => this.setState({ textSearchIndex, loading: false }),
                    error => this.setState({ error, loading: false })
                )
        )
        this.updates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <>
                <PageTitle title="Index settings" />
                <PageHeader
                    path={[{ text: 'Indexing' }]}
                    headingElement="h2"
                    className="mb-3"
                    description={
                        !this.state.error && !this.state.loading && this.state.textSearchIndex ? (
                            <>Index branches to enhance search performance at scale.</>
                        ) : undefined
                    }
                />
                <Container>
                    {this.state.loading && <LoadingSpinner />}
                    {this.state.error && (
                        <ErrorAlert prefix="Error getting repository index status" error={this.state.error} />
                    )}
                    {!this.state.error &&
                        !this.state.loading &&
                        (this.state.textSearchIndex ? (
                            <>
                                {this.state.textSearchIndex.refs && (
                                    <ul className={styles.refs}>
                                        {this.state.textSearchIndex.refs.map((reference, index) => (
                                            <TextSearchIndexedReference
                                                key={index}
                                                repo={this.props.repo}
                                                indexedRef={reference}
                                            />
                                        ))}
                                    </ul>
                                )}
                                {this.state.textSearchIndex.status && (
                                    <>
                                        <Typography.H3>Statistics</Typography.H3>
                                        <table className={classNames('table mb-0', styles.stats)}>
                                            <tbody>
                                                <tr>
                                                    <th>Last updated</th>
                                                    <td>
                                                        <Timestamp date={this.state.textSearchIndex.status.updatedAt} />
                                                    </td>
                                                </tr>
                                                <tr>
                                                    <th>Content size</th>
                                                    <td>
                                                        {prettyBytesBigint(
                                                            BigInt(this.state.textSearchIndex.status.contentByteSize)
                                                        )}{' '}
                                                        ({this.state.textSearchIndex.status.contentFilesCount}{' '}
                                                        {pluralize(
                                                            'file',
                                                            this.state.textSearchIndex.status.contentFilesCount
                                                        )}
                                                        )
                                                    </td>
                                                </tr>
                                                <tr>
                                                    <th>Index size</th>
                                                    <td>
                                                        {prettyBytes(this.state.textSearchIndex.status.indexByteSize)} (
                                                        {this.state.textSearchIndex.status.indexShardsCount}{' '}
                                                        {pluralize(
                                                            'shard',
                                                            this.state.textSearchIndex.status.indexShardsCount
                                                        )}
                                                        )
                                                    </td>
                                                </tr>
                                                <tr>
                                                    <th>New lines count</th>
                                                    <td>
                                                        {this.state.textSearchIndex.status.newLinesCount.toLocaleString()}{' '}
                                                        (default branch:{' '}
                                                        {this.state.textSearchIndex.status.defaultBranchNewLinesCount.toLocaleString()}
                                                        ) (other branches:{' '}
                                                        {this.state.textSearchIndex.status.otherBranchesNewLinesCount.toLocaleString()}
                                                        )
                                                    </td>
                                                </tr>
                                            </tbody>
                                        </table>
                                    </>
                                )}
                            </>
                        ) : (
                            <Alert className="mb-0" variant="info">
                                This Sourcegraph site has not enabled indexed search. See{' '}
                                <Link to="/help/admin/search">search documentation</Link> for information on how to
                                enable it.
                            </Alert>
                        ))}
                </Container>
            </>
        )
    }
}
