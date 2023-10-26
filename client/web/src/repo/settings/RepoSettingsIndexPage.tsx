import React from 'react'

import { mdiCheckCircle } from '@mdi/js'
import classNames from 'classnames'
import prettyBytes from 'pretty-bytes'
import { type Observable, Subject, Subscription } from 'rxjs'
import { map, switchMap, tap } from 'rxjs/operators'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { createAggregateError, pluralize } from '@sourcegraph/common'
import { gql, useMutation } from '@sourcegraph/http-client'
import {
    Button,
    Container,
    PageHeader,
    LoadingSpinner,
    Link,
    Alert,
    Icon,
    Code,
    H3,
    ErrorAlert,
    LinkOrSpan,
} from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import type {
    reindexResult,
    reindexVariables,
    RepositoryTextSearchIndexRepository,
    Scalars,
    SettingsAreaRepositoryFields,
    RepositoryTextSearchIndexResult,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { prettyBytesBigint } from '../../util/prettyBytesBigint'

import { BaseActionContainer } from './components/ActionContainer'

import styles from './RepoSettingsIndexPage.module.scss'

type RepositoryTextSearchIndex = RepositoryTextSearchIndexRepository['textSearchIndex']
/**
 * Fetches a repository's text search index information.
 */
function fetchRepositoryTextSearchIndex(id: Scalars['ID']): Observable<RepositoryTextSearchIndex> {
    return queryGraphQL<RepositoryTextSearchIndexResult>(
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
                    host {
                        name
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(({ data, errors }) => {
            if (!data?.node || errors) {
                throw createAggregateError(errors)
            }
            return (data.node as RepositoryTextSearchIndexRepository).textSearchIndex
        })
    )
}

const Reindex: React.FunctionComponent<React.PropsWithChildren<{ id: Scalars['ID'] }>> = ({ id }) => {
    const [error, setError] = React.useState<Error | null>(null)
    const [success, setSuccess] = React.useState<boolean>(false)
    const [loading, setLoading] = React.useState<boolean>(false)

    const useForceReindex = (id: Scalars['ID']): (() => void) => {
        const [submitForceReindex] = useMutation<reindexResult, reindexVariables>(
            gql`
                mutation reindex($id: ID!) {
                    reindexRepository(repository: $id) {
                        alwaysNil
                    }
                }
            `
        )
        const forceReindex = React.useCallback(() => {
            submitForceReindex({
                variables: { id },
            }).then(
                () => {
                    setLoading(false)
                    setSuccess(true)
                },
                error => {
                    setLoading(false)
                    setError(error)
                }
            )
        }, [submitForceReindex, id])
        return forceReindex
    }

    const forceReindex = useForceReindex(id)

    return (
        <BaseActionContainer
            title="Trigger Reindex"
            description={<span>Send a request to Zoekt indexserver and force an immediate reindex.</span>}
            action={
                <Button
                    variant="primary"
                    onClick={() => {
                        setLoading(true)
                        setError(null)
                        setSuccess(false)
                        forceReindex()
                    }}
                >
                    Reindex now
                </Button>
            }
            details={
                <>
                    {error && <ErrorAlert className="mt-4 mb-0" error={error} />}
                    {loading && (
                        <Alert className="mt-4 mb-0" variant="primary">
                            <LoadingSpinner /> Triggering reindex ...
                        </Alert>
                    )}
                    {success && (
                        <Alert className="mt-4 mb-0" variant="success">
                            Reindex triggered
                        </Alert>
                    )}
                </>
            }
            className="mt-0 mb-3"
        />
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
            </LinkOrSpan>
            &nbsp;&mdash;&nbsp;
            {indexedRef.indexed ? (
                <span>
                    {indexedRef.current ? 'up to date.' : 'index update in progress.'}
                    {' Last indexing job ran at '}
                    <Code>
                        <LinkOrSpan
                            to={indexedRef.indexedCommit?.commit ? indexedRef.indexedCommit.commit.url : repo.url}
                        >
                            {indexedRef.indexedCommit!.abbreviatedOID}
                        </LinkOrSpan>
                    </Code>
                    {indexedRef.skippedIndexed && Number(indexedRef.skippedIndexed.count) > 0 ? (
                        <span>
                            {', with '}
                            <Link to={'/search?q=' + encodeURIComponent(indexedRef.skippedIndexed.query)}>
                                {indexedRef.skippedIndexed.count} skipped{' '}
                                {pluralize('file', Number(indexedRef.skippedIndexed.count))}
                            </Link>
                            .
                        </span>
                    ) : null}
                </span>
            ) : (
                <span>initial indexing in progress.</span>
            )}
        </li>
    )
}

interface Props {
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
                                <Reindex id={this.props.repo.id} />
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
                                        <table className={classNames('table mb-3', styles.stats)}>
                                            <tbody>
                                                <tr>
                                                    <th>Last indexed at</th>
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
                                <>
                                    <H3>Indexserver</H3>
                                    {this.state.textSearchIndex.host ? (
                                        <table className={classNames('table mb-0', styles.stats)}>
                                            <tbody>
                                                <tr>
                                                    <th>Hostname</th>
                                                    <td>{this.state.textSearchIndex.host.name}</td>
                                                </tr>
                                            </tbody>
                                        </table>
                                    ) : (
                                        <Alert className="mb-0" variant="info">
                                            We were unable to determine the indexserver that hosts the index. However,
                                            this does not impact indexed search. The root cause is most likely a
                                            limitation of the runtime environment.
                                        </Alert>
                                    )}
                                </>
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
