import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject } from 'rxjs'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'
import { CodeIntelState } from '../shared/CodeIntelState'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../shared/CodeIntelUploadOrIndexRoot'

import { enqueueIndexJob as defaultEnqueueIndexJob, fetchLsifIndexes as defaultFetchLsifIndexes } from './backend'

export interface CodeIntelIndexesPageProps extends RouteComponentProps<{}>, TelemetryProps {
    repo?: { id: string }
    fetchLsifIndexes?: typeof defaultFetchLsifIndexes
    enqueueIndexJob?: typeof defaultEnqueueIndexJob
    now?: () => Date
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'Index state',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all indexes',
                args: {},
            },
            {
                label: 'Completed',
                value: 'completed',
                tooltip: 'Show completed indexes only',
                args: { state: LSIFIndexState.COMPLETED },
            },
            {
                label: 'Errored',
                value: 'errored',
                tooltip: 'Show errored indexes only',
                args: { state: LSIFIndexState.ERRORED },
            },
            {
                label: 'Processing',
                value: 'processing',
                tooltip: 'Show processing indexes only',
                args: { state: LSIFIndexState.PROCESSING },
            },
            {
                label: 'Queued',
                value: 'queued',
                tooltip: 'Show queued indexes only',
                args: { state: LSIFIndexState.QUEUED },
            },
        ],
    },
]

export const CodeIntelIndexesPage: FunctionComponent<CodeIntelIndexesPageProps> = ({
    repo,
    fetchLsifIndexes = defaultFetchLsifIndexes,
    enqueueIndexJob = defaultEnqueueIndexJob,
    now,
    telemetryService,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexes'), [telemetryService])

    const queryIndexes = useCallback(
        (args: FilteredConnectionQueryArguments) => fetchLsifIndexes({ repository: repo?.id, ...args }),
        [repo?.id, fetchLsifIndexes]
    )

    const querySubject = useMemo(() => new Subject<string>(), [])

    return (
        <div className="code-intel-indexes">
            <PageTitle title="Auto-indexing jobs" />
            <PageHeader headingElement="h2" path={[{ text: 'Auto-indexing jobs' }]} className="mb-3" />

            {repo && (
                <Container className="mb-2">
                    <EnqueueForm repoId={repo.id} querySubject={querySubject} enqueueIndexJob={enqueueIndexJob} />
                </Container>
            )}

            <Container>
                <div className="list-group position-relative">
                    <FilteredConnection<LsifIndexFields, Omit<CodeIntelIndexNodeProps, 'node'>>
                        listComponent="div"
                        listClassName="codeintel-indexes__grid mb-3"
                        noun="index"
                        pluralNoun="indexes"
                        querySubject={querySubject}
                        nodeComponent={CodeIntelIndexNode}
                        nodeComponentProps={{ now }}
                        queryConnection={queryIndexes}
                        history={props.history}
                        location={props.location}
                        cursorPaging={true}
                        filters={filters}
                    />
                </div>
            </Container>
        </div>
    )
}

interface CodeIntelIndexNodeProps {
    node: LsifIndexFields
    now?: () => Date
}

const CodeIntelIndexNode: FunctionComponent<CodeIntelIndexNodeProps> = ({ node, now }) => (
    <>
        <span className="codeintel-index-node__separator" />

        <div className="d-flex flex-column codeintel-index-node__information">
            <div className="m-0">
                <h3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </h3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
                </span>

                <small className="text-mute">
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, uploadedAt: null }} now={now} />
                </small>
            </div>
        </div>

        <span className="d-none d-md-inline codeintel-index-node__state">
            <CodeIntelState node={node} className="d-flex flex-column align-items-center" />
        </span>
        <span>
            <Link to={`./indexes/${node.id}`}>
                <ChevronRightIcon />
            </Link>
        </span>
    </>
)

enum State {
    Idle,
    Queueing,
    Queued,
}

interface EnqueueFormProps {
    repoId: string
    querySubject: Subject<string>
    enqueueIndexJob: typeof defaultEnqueueIndexJob
}

const EnqueueForm: FunctionComponent<EnqueueFormProps> = ({ repoId, querySubject, enqueueIndexJob }) => {
    const [revlike, setRevlike] = useState('HEAD')
    const [state, setState] = useState(() => State.Idle)
    const [queueResult, setQueueResult] = useState<number>()
    const [enqueueError, setEnqueueError] = useState<Error>()

    const enqueue = useCallback(async () => {
        setState(State.Queueing)
        setEnqueueError(undefined)
        setQueueResult(undefined)

        try {
            const indexes = await enqueueIndexJob(repoId, revlike).toPromise()
            setQueueResult(indexes.length)
            if (indexes.length > 0) {
                querySubject.next(indexes[0].inputCommit)
            }
        } catch (error) {
            setEnqueueError(error)
            setQueueResult(undefined)
        } finally {
            setState(State.Queued)
        }
    }, [repoId, revlike, querySubject, enqueueIndexJob])

    return (
        <>
            {enqueueError && <ErrorAlert prefix="Error enqueueing index job" error={enqueueError} />}

            <div className="form-inline">
                <label htmlFor="revlike">Git revlike</label>

                <input
                    type="text"
                    id="revlike"
                    className="form-control ml-2"
                    value={revlike}
                    onChange={event => setRevlike(event.target.value)}
                />

                <button
                    type="button"
                    title="Enqueue thing"
                    disabled={state === State.Queueing}
                    className="btn btn-primary ml-2"
                    onClick={enqueue}
                >
                    Enqueue
                </button>
            </div>

            {state === State.Queued && queueResult !== undefined && (
                <div className="alert alert-success mt-3 mb-0">{queueResult} index jobs enqueued.</div>
            )}
        </>
    )
}
