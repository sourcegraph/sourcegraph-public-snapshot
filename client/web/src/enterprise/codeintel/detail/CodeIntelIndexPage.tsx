import { isArray } from 'lodash'
import CheckIcon from 'mdi-react/CheckIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { timer } from 'rxjs'
import { catchError, concatMap, delay, repeatWhen, takeWhile } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { LSIFIndexState } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { ExecutionLogEntry } from '../../../components/ExecutionLogEntry'
import { PageTitle } from '../../../components/PageTitle'
import { Timestamp } from '../../../components/time/Timestamp'
import { Timeline, TimelineStage } from '../../../components/Timeline'
import { LsifIndexFields } from '../../../graphql-operations'
import { CodeIntelState } from '../shared/CodeIntelState'
import { CodeIntelStateBanner } from '../shared/CodeIntelStateBanner'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../shared/CodeIntelUploadOrIndexRoot'

import { deleteLsifIndex as defaultDeleteLsifIndex, fetchLsifIndex as defaultFetchLsifIndex } from './backend'

export interface CodeIntelIndexPageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {
    fetchLsifIndex?: typeof defaultFetchLsifIndex
    deleteLsifIndex?: typeof defaultDeleteLsifIndex
    now?: () => Date
}

const REFRESH_INTERVAL_MS = 5000

const classNamesByState = new Map([
    [LSIFIndexState.COMPLETED, 'alert-success'],
    [LSIFIndexState.ERRORED, 'alert-danger'],
])

export const CodeIntelIndexPage: FunctionComponent<CodeIntelIndexPageProps> = ({
    match: {
        params: { id },
    },
    fetchLsifIndex = defaultFetchLsifIndex,
    deleteLsifIndex = defaultDeleteLsifIndex,
    telemetryService,
    now,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndex'), [telemetryService])

    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()

    const indexOrError = useObservable(
        useMemo(
            () =>
                timer(0, REFRESH_INTERVAL_MS, undefined).pipe(
                    concatMap(() =>
                        fetchLsifIndex({ id }).pipe(
                            catchError((error): [ErrorLike] => [asError(error)]),
                            repeatWhen(observable => observable.pipe(delay(REFRESH_INTERVAL_MS)))
                        )
                    ),
                    takeWhile(shouldReload, true)
                ),
            [id, fetchLsifIndex]
        )
    )

    const deleteIndex = useCallback(async (): Promise<void> => {
        if (!indexOrError || isErrorLike(indexOrError)) {
            return
        }

        if (!window.confirm(`Delete auto-index record for commit ${indexOrError.inputCommit.slice(0, 7)}?`)) {
            return
        }

        setDeletionOrError('loading')

        try {
            await deleteLsifIndex({ id }).toPromise()
            setDeletionOrError('deleted')
        } catch (error) {
            setDeletionOrError(error)
        }
    }, [id, indexOrError, deleteLsifIndex])

    return deletionOrError === 'deleted' ? (
        <Redirect to="." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF index record" error={deletionOrError} />
    ) : (
        <div className="site-admin-lsif-index-page w-100">
            <PageTitle title="Auto-indexing jobs" />
            {isErrorLike(indexOrError) ? (
                <ErrorAlert prefix="Error loading LSIF index" error={indexOrError} />
            ) : !indexOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    <PageHeader
                        headingElement="h2"
                        path={[
                            {
                                text: `Auto-index record for ${indexOrError.projectRoot?.repository.name || ''}@${
                                    indexOrError.projectRoot
                                        ? indexOrError.projectRoot.commit.abbreviatedOID
                                        : indexOrError.inputCommit.slice(0, 7)
                                }`,
                            },
                        ]}
                        className="mb-3"
                    />

                    <Container>
                        <CodeIntelIndexMeta node={indexOrError} now={now} />
                    </Container>

                    <Container className="mt-2">
                        <CodeIntelStateBanner
                            state={indexOrError.state}
                            placeInQueue={indexOrError.placeInQueue}
                            failure={indexOrError.failure}
                            typeName="index"
                            pluralTypeName="indexes"
                            className={classNamesByState.get(indexOrError.state)}
                        />
                    </Container>

                    <Container className="mt-2">
                        <CodeIntelDeleteIndex deleteIndex={deleteIndex} deletionOrError={deletionOrError} />
                    </Container>

                    <Container className="mt-2">
                        <h3>Timeline</h3>
                        <CodeIntelIndexTimeline index={indexOrError} now={now} className="mb-3" />
                        <CodeIntelAssociatedUpload node={indexOrError} now={now} />
                    </Container>
                </>
            )}
        </div>
    )
}

const terminalStates = new Set([LSIFIndexState.COMPLETED, LSIFIndexState.ERRORED])

function shouldReload(index: LsifIndexFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(index) && !(index && terminalStates.has(index.state))
}

interface CodeIntelIndexMetaProps {
    node: LsifIndexFields
    now?: () => Date
}

const CodeIntelIndexMeta: FunctionComponent<CodeIntelIndexMetaProps> = ({ node, now }) => (
    <div className="card">
        <div className="card-body">
            <div className="card border-0">
                <div className="card-body">
                    <h3 className="card-title">
                        <CodeIntelUploadOrIndexRepository node={node} />
                    </h3>

                    <p className="card-subtitle mb-2 text-muted">
                        <CodeIntelUploadOrIndexLastActivity node={{ ...node, uploadedAt: null }} now={now} />
                    </p>

                    <p className="card-text">
                        Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                        <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
                    </p>
                </div>
            </div>
        </div>
    </div>
)

interface CodeIntelIndexTimelineProps {
    index: LsifIndexFields
    now?: () => Date
    className?: string
}

const CodeIntelIndexTimeline: FunctionComponent<CodeIntelIndexTimelineProps> = ({ index, now, className }) => {
    const stages = useMemo(
        () => [
            { icon: <TimerSandIcon />, text: 'Queued', date: index.queuedAt, className: 'bg-success' },
            { icon: <CheckIcon />, text: 'Began processing', date: index.startedAt, className: 'bg-success' },

            indexSetupStage(index, now),
            indexPreIndexStage(index, now),
            indexIndexStage(index, now),
            indexUploadStage(index, now),
            indexTeardownStage(index, now),

            index.state === LSIFIndexState.COMPLETED
                ? { icon: <CheckIcon />, text: 'Finished', date: index.finishedAt, className: 'bg-success' }
                : { icon: <ErrorIcon />, text: 'Failed', date: index.finishedAt, className: 'bg-danger' },
        ],
        [index, now]
    )

    return <Timeline stages={stages.filter(isDefined)} now={now} className={className} />
}

const indexSetupStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined =>
    index.steps.setup.length === 0
        ? undefined
        : {
              text: 'Setup',
              details: index.steps.setup.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(index.steps.setup),
          }

const indexPreIndexStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined => {
    const logEntries = index.steps.preIndex.map(step => step.logEntry).filter(isDefined)

    return logEntries.length === 0
        ? undefined
        : {
              text: 'Pre Index',
              details: index.steps.preIndex.map(
                  step =>
                      step.logEntry && (
                          <div key={`${step.image}${step.root}${step.commands.join(' ')}}`}>
                              <ExecutionLogEntry logEntry={step.logEntry} now={now}>
                                  <ExecutionMetaInformation
                                      {...{
                                          image: step.image,
                                          commands: step.commands,
                                          root: step.root,
                                      }}
                                  />
                              </ExecutionLogEntry>
                          </div>
                      )
              ),
              ...genericStage(logEntries),
          }
}

const indexIndexStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined =>
    !index.steps.index.logEntry
        ? undefined
        : {
              text: 'Index',
              details: (
                  <>
                      <ExecutionLogEntry logEntry={index.steps.index.logEntry} now={now}>
                          <ExecutionMetaInformation
                              {...{
                                  image: index.inputIndexer,
                                  commands: index.steps.index.indexerArgs,
                                  root: index.inputRoot,
                              }}
                          />
                      </ExecutionLogEntry>
                  </>
              ),
              ...genericStage(index.steps.index.logEntry),
          }

const indexUploadStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined =>
    !index.steps.upload
        ? undefined
        : {
              text: 'Upload',
              details: <ExecutionLogEntry logEntry={index.steps.upload} now={now} />,
              ...genericStage(index.steps.upload),
          }

const indexTeardownStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined =>
    index.steps.teardown.length === 0
        ? undefined
        : {
              text: 'Teardown',
              details: index.steps.teardown.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(index.steps.teardown),
          }

const genericStage = <E extends { startTime: string; exitCode: number | null }>(
    value: E | E[]
): Pick<TimelineStage, 'icon' | 'date' | 'className' | 'expanded'> => {
    const finished = isArray(value) ? value.every(logEntry => logEntry.exitCode !== null) : value.exitCode !== null
    const success = isArray(value) ? value.every(logEntry => logEntry.exitCode === 0) : value.exitCode === 0

    return {
        icon: !finished ? <ProgressClockIcon /> : success ? <CheckIcon /> : <ErrorIcon />,
        date: isArray(value) ? value[0].startTime : value.startTime,
        className: success || !finished ? 'bg-success' : 'bg-danger',
        expanded: !(success || !finished),
    }
}

const ExecutionMetaInformation: React.FunctionComponent<{ image: string; commands: string[]; root: string }> = ({
    image,
    commands,
    root,
}) => (
    <div className="pt-3">
        <div className="docker-command-spec py-2 border-top pl-2">
            <strong className="docker-command-spec__header">Image</strong>
            <div>{image}</div>
        </div>
        <div className="docker-command-spec py-2 border-top pl-2">
            <strong className="docker-command-spec__header">Commands</strong>
            <div>
                <code>{commands.join(' ')}</code>
            </div>
        </div>
        <div className="docker-command-spec py-2 border-top pl-2">
            <strong className="docker-command-spec__header">Root</strong>
            <div>/{root}</div>
        </div>
    </div>
)

const CodeIntelAssociatedUpload: FunctionComponent<CodeIntelAssociatedUploadProps> = ({ node, now }) =>
    node.associatedUpload && node.projectRoot ? (
        <>
            <div className="list-group position-relative">
                <div className="codeintel-associated-upload__grid">
                    <span className="codeintel-associated-upload__separator" />

                    <div className="d-flex flex-column codeintel-associated-upload__information">
                        <div className="m-0">
                            <h3 className="m-0 d-block d-md-inline">
                                This job uploaded an index{' '}
                                <Timestamp date={node.associatedUpload.uploadedAt} now={now} />
                            </h3>
                        </div>

                        <div>
                            <small className="text-mute">
                                <CodeIntelUploadOrIndexLastActivity
                                    node={{ ...node.associatedUpload, queuedAt: null }}
                                    now={now}
                                />
                            </small>
                        </div>
                    </div>

                    <span className="d-none d-md-inline codeintel-associated-upload__state">
                        <CodeIntelState
                            node={node.associatedUpload}
                            className="d-flex flex-column align-items-center"
                        />
                    </span>
                    <span>
                        <Link
                            to={`/${node.projectRoot.repository.name}/-/settings/code-intelligence/uploads/${node.associatedUpload.id}`}
                        >
                            <ChevronRightIcon />
                        </Link>
                    </span>
                </div>
            </div>
        </>
    ) : (
        <></>
    )

interface CodeIntelDeleteIndexProps {
    deleteIndex: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

const CodeIntelDeleteIndex: FunctionComponent<CodeIntelDeleteIndexProps> = ({ deleteIndex, deletionOrError }) => (
    <button
        type="button"
        className="btn btn-outline-danger"
        onClick={deleteIndex}
        disabled={deletionOrError === 'loading'}
        aria-describedby="upload-delete-button-help"
        data-tooltip="Deleting this index will remove it from the index queue."
    >
        <DeleteIcon className="icon-inline" /> Delete index
    </button>
)

interface CodeIntelAssociatedUploadProps {
    node: LsifIndexFields
    now?: () => Date
}
