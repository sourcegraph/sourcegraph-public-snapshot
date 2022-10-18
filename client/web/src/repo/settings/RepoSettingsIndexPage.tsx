import * as React from 'react'

import { mdiCheckCircle } from '@mdi/js'
import classNames from 'classnames'
import prettyBytes from 'pretty-bytes'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { map, switchMap, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { createAggregateError, pluralize } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Container, PageHeader, LoadingSpinner, Link, Alert, Icon, Code, H3 } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { RepositoryTextSearchIndexRepository, Scalars, SettingsAreaRepositoryFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { prettyBytesBigint } from '../../util/prettyBytesBigint'

import styles from './RepoSettingsIndexPage.module.scss'

type RepositoryTextSearchIndex = RepositoryTextSearchIndexRepository['textSearchIndex']
/**
 * Fetches a repository's text search index information.
 */
function fetchRepositoryTextSearchIndex(id: Scalars['ID']): Observable<RepositoryTextSearchIndex> {
    return queryGraphQL(
        gql`
            query RepositoryTextSearchIndex($id: ID!) {
                node(id: $id) {
                    ...RepositoryTextSearchIndexRepository
                }
            }

            fragment RepositoryTextSearchIndexRepository on Repository {
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
                        skippedIndexed {
                            count
                            query
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
        `,
        { id }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node || errors) {
                throw createAggregateError(errors)
            }
            return (data.node as RepositoryTextSearchIndexRepository).textSearchIndex
        })
    )
}

const TextSearchIndexedReference: React.FunctionComponent<
    React.PropsWithChildren<{
        repo: SettingsAreaRepositoryFields
        indexedRef: NonNullable<RepositoryTextSearchIndex>['refs'][number]
    }>
> = ({ repo, indexedRef }) => {
    const isCurrent = indexedRef.indexed && indexedRef.current

    return (
        <li className={styles.ref}>
            <Icon
                className={classNames(styles.refIcon, isCurrent && styles.refIconCurrent)}
                svgPath={isCurrent ? mdiCheckCircle : undefined}
                as={!isCurrent ? LoadingSpinner : undefined}
                aria-hidden={true}
            />
            <LinkOrSpan to={indexedRef.ref.url}>
                <Code weight="bold">{indexedRef.ref.displayName}</Code>
            </LinkOrSpan>{' '}
            {indexedRef.indexed ? (
                <span>
                    &nbsp;&mdash; indexed at{' '}
                    <Code>
                        <LinkOrSpan
                            to={indexedRef.indexedCommit?.commit ? indexedRef.indexedCommit.commit.url : repo.url}
                        >
                            {indexedRef.indexedCommit!.abbreviatedOID}
                        </LinkOrSpan>
                    </Code>{' '}
                    {indexedRef.current ? '(up to date)' : '(index update in progress)'}
                    {indexedRef.skippedIndexed && Number(indexedRef.skippedIndexed.count) > 0 ? (
                        <span>
                            .&nbsp;
                            <Link to={'/search?q=' + encodeURIComponent(indexedRef.skippedIndexed.query)}>
                                {indexedRef.skippedIndexed.count}{' '}
                                {pluralize('file', Number(indexedRef.skippedIndexed.count))} skipped during indexing
                            </Link>
                            .
                        </span>
                    ) : null}
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
    textSearchIndex?: RepositoryTextSearchIndex
    loading: boolean
    error?: Error
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
                                        <H3>Statistics</H3>
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
                                                        )}
                                                    </td>
                                                </tr>
                                                <tr>
                                                    <th>Shards</th>
                                                    <td>{this.state.textSearchIndex.status.indexShardsCount}</td>
                                                </tr>
                                                <tr>
                                                    <th>Files</th>
                                                    <td>{this.state.textSearchIndex.status.contentFilesCount}</td>
                                                </tr>
                                                <tr>
                                                    <th>Index size</th>
                                                    <td>
                                                        {prettyBytes(this.state.textSearchIndex.status.indexByteSize)}
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
