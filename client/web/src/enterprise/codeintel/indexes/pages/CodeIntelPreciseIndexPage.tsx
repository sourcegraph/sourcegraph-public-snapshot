import { FunctionComponent, ReactNode, useCallback, useEffect, useMemo, useState } from 'react'

import { Redirect, RouteComponentProps, useLocation } from 'react-router'

import { useApolloClient } from '@apollo/client'
import {
    mdiDatabaseEdit,
    mdiDatabasePlus,
    mdiDelete,
    mdiGraph,
    mdiHistory,
    mdiInformationOutline,
    mdiMapSearch,
    mdiRecycle,
    mdiRedo,
    mdiTimerSand,
} from '@mdi/js'
import { ErrorLike, isErrorLike, pluralize } from '@sourcegraph/common'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    Alert,
    AlertProps,
    Button,
    Card,
    CardBody,
    CardText,
    CardTitle,
    Container,
    ErrorAlert,
    ErrorMessage,
    H3,
    Icon,
    Link,
    LoadingSpinner,
    PageHeader,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
    Text,
    Tooltip,
    useObservable,
} from '@sourcegraph/wildcard'
import classNames from 'classnames'
import * as H from 'history'
import { Observable } from 'rxjs'
import { takeWhile } from 'rxjs/operators'
import {
    Connection,
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { Timeline, TimelineStage } from '../../../../components/Timeline'
import {
    AuditLogOperation,
    LsifUploadsAuditLogsFields,
    PreciseIndexFields,
    PreciseIndexState,
} from '../../../../graphql-operations'
import { CodeIntelUploadOrIndexCommit } from '../../shared/components/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexCommitTags } from '../../shared/components/CodeIntelUploadOrIndexCommitTags'
import { CodeIntelUploadOrIndexRepository } from '../../shared/components/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../../shared/components/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexRoot } from '../../shared/components/CodeIntelUploadOrIndexRoot'
import { PreciseIndexLastUpdated } from '../components/CodeIntelLastUpdated'
import { IndexTimeline } from '../components/IndexTimeline'
import { ProjectDescription } from '../components/ProjectDescriptionProps'
import { queryDependencyGraph } from '../hooks/queryDependencyGraph'
import { queryPreciseIndex } from '../hooks/queryPreciseIndex'
import {
    NormalizedUploadRetentionMatch,
    queryPreciseIndexRetention,
    RetentionPolicyMatch,
    UploadReferenceMatch,
} from '../hooks/queryPreciseIndexRetention'
import styles from './CodeIntelPreciseIndexPage.module.scss'
import { useDeletePreciseIndex } from '../hooks/useDeletePreciseIndex'
import { useReindexPreciseIndex } from '../hooks/useReindexPreciseIndex'
import { FlashMessage } from '../../configuration/components/FlashMessage'

export interface CodeIntelPreciseIndexPageProps
    extends RouteComponentProps<{ id: string }>,
        ThemeProps,
        TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    now?: () => Date
}

enum RetentionPolicyMatcherState {
    ShowMatchingOnly,
    ShowAll,
}

const variantByState = new Map<PreciseIndexState, AlertProps['variant']>([
    [PreciseIndexState.COMPLETED, 'success'],
    [PreciseIndexState.INDEXING_ERRORED, 'danger'],
    [PreciseIndexState.PROCESSING_ERRORED, 'danger'],
])

export const CodeIntelPreciseIndexPage: FunctionComponent<CodeIntelPreciseIndexPageProps> = ({
    match: {
        params: { id },
    },
    authenticatedUser,
    now,
    history,
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelPreciseIndexPage'), [telemetryService])
    const location = useLocation<{ message: string; modal: string }>()

    const apolloClient = useApolloClient()
    const [reindexOrError, setReindexOrError] = useState<'loading' | 'reindexed' | ErrorLike>()
    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()
    const { handleDeletePreciseIndex, deleteError } = useDeletePreciseIndex()
    const { handleReindexPreciseIndex, reindexError } = useReindexPreciseIndex()
    const [retentionPolicyMatcherState, setRetentionPolicyMatcherState] = useState(RetentionPolicyMatcherState.ShowAll)

    const indexOrError = useObservable(
        useMemo(
            () => queryPreciseIndex(id, apolloClient).pipe(takeWhile(shouldReload, true)),
            [id, queryPreciseIndex, apolloClient]
        )
    )

    useEffect(() => {
        if (deleteError) {
            setDeletionOrError(deleteError)
        }
    }, [deleteError])

    useEffect(() => {
        if (reindexError) {
            setReindexOrError(reindexError)
        }
    }, [reindexError])

    const reindexUpload = useCallback(async (): Promise<void> => {
        if (!indexOrError || isErrorLike(indexOrError)) {
            return
        }

        setReindexOrError('loading')

        try {
            await handleReindexPreciseIndex({
                variables: { id },
                update: cache => cache.modify({ fields: { node: () => {} } }),
            })
            setReindexOrError('reindexed')
            history.push({
                state: {
                    modal: 'SUCCESS',
                    message: `Marked as replaceable.`, // TODO
                },
            })
        } catch (error) {
            setReindexOrError(error)
            history.push({
                state: {
                    modal: 'ERROR',
                    message: `There was an error while marking index as replaceable.`, // TODO
                },
            })
        }
    }, [id, indexOrError, handleReindexPreciseIndex, history])

    const deleteUpload = useCallback(async (): Promise<void> => {
        if (!indexOrError || isErrorLike(indexOrError)) {
            return
        }

        let description = `${indexOrError.inputCommit.slice(0, 7)}`
        if (indexOrError.inputRoot) {
            description += ` rooted at ${indexOrError.inputRoot}`
        }

        if (!window.confirm(`Delete upload for commit ${description}?`)) {
            return
        }

        setDeletionOrError('loading')

        try {
            await handleDeletePreciseIndex({
                variables: { id },
                update: cache => cache.modify({ fields: { node: () => {} } }),
            })
            setDeletionOrError('deleted')
            history.push({
                state: {
                    modal: 'SUCCESS',
                    message: `Upload for commit ${description} is deleting.`,
                },
            })
        } catch (error) {
            setDeletionOrError(error)
            history.push({
                state: {
                    modal: 'ERROR',
                    message: `There was an error while deleting upload for commit ${description}.`,
                },
            })
        }
    }, [id, indexOrError, handleDeletePreciseIndex, history])

    const queryDependents = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (indexOrError && !isErrorLike(indexOrError)) {
                return queryDependencyGraph({ ...args, dependentOf: indexOrError.id }, apolloClient)
            }

            throw new Error('unreachable: queryDependents referenced with invalid upload')
        },
        [indexOrError, queryDependencyGraph, apolloClient]
    )

    const queryRetentionPoliciesCallback = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<Connection<NormalizedUploadRetentionMatch>> => {
            if (indexOrError && !isErrorLike(indexOrError)) {
                return queryPreciseIndexRetention(apolloClient, id, {
                    matchesOnly: retentionPolicyMatcherState === RetentionPolicyMatcherState.ShowMatchingOnly,
                    ...args,
                })
            }

            throw new Error('unreachable: queryRetentionPolicies referenced with invalid upload')
        },
        [indexOrError, apolloClient, id, queryPreciseIndexRetention, retentionPolicyMatcherState]
    )

    return deletionOrError === 'deleted' ? (
        <Redirect to="." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting precise index" error={deletionOrError} />
    ) : isErrorLike(indexOrError) ? (
        <ErrorAlert prefix="Error fetching precise index" error={indexOrError} />
    ) : !indexOrError ? (
        <LoadingSpinner />
    ) : (
        <>
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: (
                            <>
                                Precise index of{' '}
                                {indexOrError.projectRoot
                                    ? `${indexOrError.projectRoot.repository.name}@${indexOrError.projectRoot.commit.abbreviatedOID}`
                                    : 'an unknown commit'}
                            </>
                        ),
                    },
                ]}
                className="mb-3"
            />

            {!!location.state && <FlashMessage state={location.state.modal} message={location.state.message} />}

            <Container>
                <IndexDescription index={indexOrError} />

                <div className="mt-2">
                    <Alert variant={variantByState.get(indexOrError.state) ?? 'primary'}>
                        <span>
                            {indexOrError.state === PreciseIndexState.UPLOADING_INDEX ? (
                                <span>Still uploading...</span>
                            ) : indexOrError.state === PreciseIndexState.DELETING ? (
                                <span>Upload is queued for deletion.</span>
                            ) : indexOrError.state === PreciseIndexState.QUEUED_FOR_INDEXING ? (
                                <>
                                    Index is queued for indexing.{' '}
                                    <LousyDescription placeInQueue={indexOrError.placeInQueue} />
                                </>
                            ) : indexOrError.state === PreciseIndexState.QUEUED_FOR_PROCESSING ? (
                                <>
                                    <span>
                                        Index is queued for processing.{' '}
                                        <LousyDescription placeInQueue={indexOrError.placeInQueue} />
                                    </span>
                                </>
                            ) : indexOrError.state === PreciseIndexState.INDEXING ? (
                                <span>Index is currently being indexed...</span>
                            ) : indexOrError.state === PreciseIndexState.PROCESSING ? (
                                <span>Index is currently being processed...</span>
                            ) : indexOrError.state === PreciseIndexState.COMPLETED ? (
                                <span>Index processed successfully.</span>
                            ) : indexOrError.state === PreciseIndexState.INDEXING_ERRORED ? (
                                <span>
                                    Index failed to index: <ErrorMessage error={indexOrError.failure} />
                                </span>
                            ) : indexOrError.state === PreciseIndexState.PROCESSING_ERRORED ? (
                                <span>
                                    Index failed to process: <ErrorMessage error={indexOrError.failure} />
                                </span>
                            ) : (
                                <></>
                            )}
                        </span>
                    </Alert>

                    {indexOrError.isLatestForRepo && (
                        <Alert variant={'secondary'}>
                            <span>
                                This upload can answer queries for the tip of the default branch and are targets of
                                cross-repository find reference operations.
                            </span>
                        </Alert>
                    )}
                </div>

                <Tabs size="medium" className={classNames('mt-2', styles.tabs)}>
                    <TabList>
                        <Tab>
                            <span>
                                <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiTimerSand} />
                                <span className="text-content" data-tab-content="Timeline">
                                    Timeline
                                </span>
                            </span>
                        </Tab>

                        {(indexOrError.state === PreciseIndexState.COMPLETED ||
                            indexOrError.state === PreciseIndexState.DELETING) && (
                            <>
                                <Tab>
                                    <span>
                                        <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiGraph} />
                                        <span className="text-content" data-tab-content="Dependencies">
                                            Dependencies
                                        </span>
                                    </span>
                                </Tab>
                                <Tab>
                                    <span>
                                        <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiGraph} />
                                        <span className="text-content" data-tab-content="Dependents">
                                            Dependents
                                        </span>
                                    </span>
                                </Tab>
                                <Tab>
                                    <span>
                                        <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiRecycle} />
                                        <span className="text-content" data-tab-content="Retention">
                                            Retention
                                        </span>
                                    </span>
                                </Tab>
                                <Tab>
                                    <span>
                                        <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiHistory} />
                                        <span className="text-content" data-tab-content="Audit logs">
                                            Audit logs
                                        </span>
                                    </span>
                                </Tab>
                            </>
                        )}
                    </TabList>
                    <TabPanels>
                        <TabPanel>
                            <Container className="mt-2">
                                <IndexTimeline index={indexOrError} />
                            </Container>
                        </TabPanel>

                        {(indexOrError.state === PreciseIndexState.COMPLETED ||
                            indexOrError.state === PreciseIndexState.DELETING) && (
                            <>
                                <TabPanel>
                                    <Container className="mt-2">
                                        <DependencyList index={indexOrError} history={history} location={location} />
                                    </Container>
                                </TabPanel>
                                <TabPanel>
                                    <Container className="mt-2">
                                        <FilteredConnection
                                            listComponent="div"
                                            listClassName={classNames(styles.grid, 'mb-3')}
                                            inputClassName="w-auto"
                                            noun="dependent"
                                            pluralNoun="dependents"
                                            nodeComponent={DependencyOrDependentNode}
                                            queryConnection={queryDependents}
                                            history={history}
                                            location={location}
                                            cursorPaging={true}
                                            useURLQuery={false}
                                            // emptyElement={<EmptyDependents />}
                                        />
                                    </Container>
                                </TabPanel>
                                <TabPanel>
                                    <Container className="mt-2">
                                        {retentionPolicyMatcherState === RetentionPolicyMatcherState.ShowAll ? (
                                            <Button
                                                type="button"
                                                className="float-right p-0 mb-2"
                                                variant="link"
                                                onClick={() =>
                                                    setRetentionPolicyMatcherState(
                                                        RetentionPolicyMatcherState.ShowMatchingOnly
                                                    )
                                                }
                                            >
                                                Show matching only
                                            </Button>
                                        ) : (
                                            <Button
                                                type="button"
                                                className="float-right p-0 mb-2"
                                                variant="link"
                                                onClick={() =>
                                                    setRetentionPolicyMatcherState(RetentionPolicyMatcherState.ShowAll)
                                                }
                                            >
                                                Show all
                                            </Button>
                                        )}
                                        <FilteredConnection
                                            listComponent="div"
                                            listClassName={classNames(styles.grid, 'mb-3')}
                                            inputClassName="w-auto"
                                            noun="match"
                                            pluralNoun="matches"
                                            nodeComponent={RetentionMatchNode}
                                            queryConnection={queryRetentionPoliciesCallback}
                                            history={history}
                                            location={location}
                                            cursorPaging={true}
                                            useURLQuery={false}
                                            emptyElement={<EmptyUploadRetentionMatchStatus />}
                                        />
                                    </Container>
                                </TabPanel>
                                <TabPanel>
                                    <Container className="mt-2">
                                        {indexOrError.auditLogs?.length ?? 0 > 0 ? (
                                            <UploadAuditLogTimeline logs={indexOrError.auditLogs || []} />
                                        ) : (
                                            <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
                                                <Icon
                                                    className="mb-2"
                                                    svgPath={mdiMapSearch}
                                                    inline={false}
                                                    aria-hidden={true}
                                                />
                                                <br />
                                                This upload has no audit logs.
                                            </Text>
                                        )}
                                    </Container>
                                </TabPanel>
                            </>
                        )}
                    </TabPanels>
                </Tabs>
            </Container>

            <Container className="mt-2">
                {authenticatedUser?.siteAdmin && (
                    <>
                        <CodeIntelDeleteUpload
                            state={indexOrError.state}
                            deleteUpload={deleteUpload}
                            deletionOrError={deletionOrError}
                        />

                        <CodeIntelReindexUpload reindexUpload={reindexUpload} reindexOrError={reindexOrError} />
                    </>
                )}
            </Container>
        </>
    )
}

const terminalStates = new Set(['TODO']) // TODO

function shouldReload(index: PreciseIndexFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(index) && !(index && terminalStates.has(index.state))
}

//
//
//

interface CodeIntelStateDescriptionPlaceInQueueProps {
    placeInQueue?: number | null
}

const LousyDescription: FunctionComponent<React.PropsWithChildren<CodeIntelStateDescriptionPlaceInQueueProps>> = ({
    placeInQueue,
}) => {
    if (placeInQueue === 1) {
        return <>This index is up next for processing.</>
    }

    return <>{placeInQueue ? `There are ${placeInQueue - 1} indexes ahead of this one.` : ''}</>
}

interface RetentionMatchNodeProps {
    node: NormalizedUploadRetentionMatch
}

const retentionByUploadTitle = 'Retention by reference'
const retentionByBranchTipTitle = 'Retention by tip of default branch'

const RetentionMatchNode: FunctionComponent<React.PropsWithChildren<RetentionMatchNodeProps>> = ({ node }) => {
    if (node.matchType === 'RetentionPolicy') {
        return <RetentionPolicyRetentionMatchNode match={node} />
    }
    if (node.matchType === 'UploadReference') {
        return <UploadReferenceRetentionMatchNode match={node} />
    }

    throw new Error(`invalid node type ${JSON.stringify(node as object)}`)
}

const RetentionPolicyRetentionMatchNode: FunctionComponent<
    React.PropsWithChildren<{ match: RetentionPolicyMatch }>
> = ({ match }) => (
    <>
        <span className={styles.separator} />

        <div className={classNames(styles.information, 'd-flex flex-column')}>
            <div className="m-0">
                {match.configurationPolicy ? (
                    <Link to={`../configuration/${match.configurationPolicy.id}`} className="p-0">
                        <H3 className="m-0 d-block d-md-inline">{match.configurationPolicy.name}</H3>
                    </Link>
                ) : (
                    <H3 className="m-0 d-block d-md-inline">{retentionByBranchTipTitle}</H3>
                )}
                <div className="mr-2 d-block d-mdinline-block">
                    Retained: {match.matches ? 'yes' : 'no'}
                    {match.protectingCommits.length !== 0 && (
                        <>
                            , by {match.protectingCommits.length} visible{' '}
                            {pluralize('commit', match.protectingCommits.length)}, including{' '}
                            {match.protectingCommits
                                .slice(0, 4)
                                .map(hash => hash.slice(0, 7))
                                .join(', ')}
                            <Tooltip content="This upload is retained to service code-intel queries for commit(s) with applicable retention policies.">
                                <Icon
                                    aria-label="This upload is retained to service code-intel queries for commit(s) with applicable retention policies."
                                    className="ml-1"
                                    svgPath={mdiInformationOutline}
                                />
                            </Tooltip>
                        </>
                    )}
                    {!match.configurationPolicy && (
                        <Tooltip content="Uploads at the tip of the default branch are always retained indefinitely.">
                            <Icon
                                aria-label="Uploads at the tip of the default branch are always retained indefinitely."
                                className="ml-1"
                                svgPath={mdiInformationOutline}
                            />
                        </Tooltip>
                    )}
                </div>
            </div>
        </div>
    </>
)

const UploadReferenceRetentionMatchNode: FunctionComponent<
    React.PropsWithChildren<{ match: UploadReferenceMatch }>
> = ({ match }) => (
    <>
        <span className={styles.separator} />

        <div className={classNames(styles.information, 'd-flex flex-column')}>
            <div className="m-0">
                <H3 className="m-0 d-block d-md-inline">{retentionByUploadTitle}</H3>
                <div className="mr-2 d-block d-mdinline-block">
                    Referenced by {match.total} {pluralize('upload', match.total, 'uploads')}, including{' '}
                    {match.uploadSlice
                        .slice(0, 3)
                        .map<React.ReactNode>(upload => (
                            <Link key={upload.id} to={`/site-admin/code-graph/uploads/${upload.id}`}>
                                {upload.projectRoot?.repository.name ?? 'unknown'}
                            </Link>
                        ))
                        .reduce((previous, current) => [previous, ', ', current])}
                    <Tooltip content="Uploads that are dependencies of other upload(s) are retained to service cross-repository code-intel queries.">
                        <Icon
                            aria-label="Uploads that are dependencies of other upload(s) are retained to service cross-repository code-intel queries."
                            className="ml-1"
                            svgPath={mdiInformationOutline}
                        />
                    </Tooltip>
                </div>
            </div>
        </div>
    </>
)

interface DependencyOrDependentNodeProps {
    node: PreciseIndexFields
    now?: () => Date
}

const DependencyOrDependentNode: FunctionComponent<React.PropsWithChildren<DependencyOrDependentNodeProps>> = ({
    node,
}) => (
    <>
        <span className={styles.separator} />

        <div className={classNames(styles.information, 'd-flex flex-column')}>
            <div className="m-0">
                <H3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </H3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} />
                    {node.tags.length > 0 && (
                        <>
                            , <CodeIntelUploadOrIndexCommitTags tags={node.tags} />,
                        </>
                    )}{' '}
                    by <CodeIntelUploadOrIndexIndexer node={node} />
                </span>
            </div>
        </div>
    </>
)

interface UploadAuditLogTimelineProps {
    logs: LsifUploadsAuditLogsFields[]
}

const UploadAuditLogTimeline: FunctionComponent<React.PropsWithChildren<UploadAuditLogTimelineProps>> = ({ logs }) => {
    const stages = logs?.map(
        (log): TimelineStage => ({
            icon:
                log.operation === AuditLogOperation.CREATE ? (
                    <Icon aria-label="Success" svgPath={mdiDatabasePlus} />
                ) : (
                    <Icon aria-label="Warn" svgPath={mdiDatabaseEdit} />
                ),
            text: stageText(log),
            className: log.operation === AuditLogOperation.CREATE ? 'bg-success' : 'bg-warning',
            expandedByDefault: true,
            date: log.logTimestamp,
            details: (
                <>
                    {log.reason && (
                        <>
                            <Container>
                                <b>Reason</b>: {log.reason}
                            </Container>
                            <br />
                        </>
                    )}
                    <div className={styles.tableContainer}>
                        <table className="table mb-0 table-striped">
                            <thead>
                                <tr>
                                    <th className={styles.dbColumnCol} scope="column">
                                        Column
                                    </th>
                                    <th className={styles.dataColumnCol} scope="column">
                                        Old
                                    </th>
                                    <th scope="column">New</th>
                                </tr>
                            </thead>
                            <tbody>
                                {log.changedColumns.map((change, index) => (
                                    // eslint-disable-next-line react/no-array-index-key
                                    <tr key={index} className="overflow-scroll">
                                        <td className="mr-2">{change.column}</td>
                                        <td className="mr-2">{change.old || 'NULL'}</td>
                                        <td className="mr-2">{change.new || 'NULL'}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </>
            ),
        })
    )

    return <Timeline showDurations={false} stages={stages} />
}

function stageText(log: LsifUploadsAuditLogsFields): ReactNode {
    if (log.operation === AuditLogOperation.CREATE) {
        return 'Upload created'
    }

    return (
        <>
            Altered columns:{' '}
            {formatReactNodeList(log.changedColumns.map(change => <span key={change.column}>{change.column}</span>))}
        </>
    )
}

function formatReactNodeList(list: ReactNode[]): ReactNode {
    if (list.length === 0) {
        return <></>
    }
    if (list.length === 1) {
        return list[0]
    }

    return (
        <>
            {list.slice(0, -1).reduce((previous, current) => [previous, ', ', current])} and {list[list.length - 1]}
        </>
    )
}

const EmptyUploadRetentionMatchStatus: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No retention policies matched.
    </Text>
)

interface CodeIntelDeleteUploadProps {
    state: PreciseIndexState
    deleteUpload: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

const CodeIntelDeleteUpload: FunctionComponent<React.PropsWithChildren<CodeIntelDeleteUploadProps>> = ({
    state,
    deleteUpload,
    deletionOrError,
}) =>
    state === PreciseIndexState.DELETING ? (
        <></>
    ) : (
        <Tooltip
            content={
                state === PreciseIndexState.COMPLETED
                    ? 'Deleting this index will make it unavailable to answer code navigation queries the next time the repository commit graph is refreshed.'
                    : 'Delete this index immediately'
            }
        >
            <Button
                type="button"
                className="float-right"
                variant="danger"
                onClick={deleteUpload}
                disabled={deletionOrError === 'loading'}
            >
                <Icon aria-hidden={true} svgPath={mdiDelete} /> Delete index
            </Button>
        </Tooltip>
    )

interface CodeIntelReindexUploadProps {
    reindexUpload: () => Promise<void>
    reindexOrError?: 'loading' | 'reindexed' | ErrorLike
}

const CodeIntelReindexUpload: FunctionComponent<React.PropsWithChildren<CodeIntelReindexUploadProps>> = ({
    reindexUpload,
    reindexOrError,
}) => (
    <Tooltip content={'TODO'}>
        <Button type="button" variant="link" onClick={reindexUpload} disabled={reindexOrError === 'loading'}>
            <Icon aria-hidden={true} svgPath={mdiRedo} /> Mark index as replaceable by autoindexing
        </Button>
    </Tooltip>
)

//
//
//

interface IndexDescriptionProps {
    index: PreciseIndexFields
}

const IndexDescription: FunctionComponent<IndexDescriptionProps> = ({ index }) => (
    <Card>
        <CardBody>
            <CardTitle>
                {index.projectRoot ? (
                    <Link to={index.projectRoot.repository.url}>{index.projectRoot.repository.name}</Link>
                ) : (
                    <span>Unknown repository</span>
                )}
            </CardTitle>

            <CardText>
                <span className="d-block">
                    <ProjectDescription index={index} />
                </span>

                <small className="text-mute">
                    <PreciseIndexLastUpdated index={index} />
                </small>
            </CardText>
        </CardBody>
    </Card>
)

export interface DependencyListProps {
    index: PreciseIndexFields
    history: H.History
    location: H.Location
}

export const DependencyList: FunctionComponent<DependencyListProps> = ({ index, history, location }) => {
    const apolloClient = useApolloClient()
    const queryDependencies = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (index && !isErrorLike(index)) {
                return queryDependencyGraph({ ...args, dependencyOf: index.id }, apolloClient)
            }
            throw new Error('unreachable: queryDependencies referenced with invalid upload')
        },
        [index, queryDependencyGraph, apolloClient]
    )

    return (
        <FilteredConnection
            listComponent="div"
            listClassName={classNames(styles.grid, 'mb-3')}
            inputClassName="w-auto"
            noun="dependency"
            pluralNoun="dependencies"
            nodeComponent={DependencyOrDependentNode}
            queryConnection={queryDependencies}
            history={history}
            location={location}
            cursorPaging={true}
            useURLQuery={false}
            // emptyElement={<EmptyDependencies />}
        />
    )
}
