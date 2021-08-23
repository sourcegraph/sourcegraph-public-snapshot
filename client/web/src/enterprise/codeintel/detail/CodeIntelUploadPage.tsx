import CheckIcon from 'mdi-react/CheckIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import FileUploadIcon from 'mdi-react/FileUploadIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { timer } from 'rxjs'
import { catchError, concatMap, delay, repeatWhen, takeWhile } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { LSIFUploadState } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import {
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { Timeline, TimelineStage } from '@sourcegraph/web/src/components/Timeline'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { LsifUploadFields } from '../../../graphql-operations'
import { fetchLsifUploads as defaultFetchLsifUploads } from '../shared/backend'
import { CodeIntelState } from '../shared/CodeIntelState'
import { CodeIntelStateBanner } from '../shared/CodeIntelStateBanner'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../shared/CodeIntelUploadOrIndexRoot'

import { deleteLsifUpload as defaultDeleteLsifUpload, fetchLsifUpload as defaultFetchUpload } from './backend'

export interface CodeIntelUploadPageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {
    fetchLsifUpload?: typeof defaultFetchUpload
    fetchLsifUploads?: typeof defaultFetchLsifUploads
    deleteLsifUpload?: typeof defaultDeleteLsifUpload
    now?: () => Date
}

const REFRESH_INTERVAL_MS = 5000

const classNamesByState = new Map([
    [LSIFUploadState.COMPLETED, 'alert-success'],
    [LSIFUploadState.ERRORED, 'alert-danger'],
])

enum DependencyGraphState {
    ShowDependencies,
    ShowDependents,
}

export const CodeIntelUploadPage: FunctionComponent<CodeIntelUploadPageProps> = ({
    match: {
        params: { id },
    },
    fetchLsifUpload = defaultFetchUpload,
    fetchLsifUploads = defaultFetchLsifUploads,
    deleteLsifUpload = defaultDeleteLsifUpload,
    telemetryService,
    now,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelUpload'), [telemetryService])

    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()
    const [dependencyGraphState, setDependencyGraphState] = useState(DependencyGraphState.ShowDependencies)

    const uploadOrError = useObservable(
        useMemo(
            () =>
                timer(0, REFRESH_INTERVAL_MS, undefined).pipe(
                    concatMap(() =>
                        fetchLsifUpload({ id }).pipe(
                            catchError((error): [ErrorLike] => [asError(error)]),
                            repeatWhen(observable => observable.pipe(delay(REFRESH_INTERVAL_MS)))
                        )
                    ),
                    takeWhile(shouldReload, true)
                ),
            [id, fetchLsifUpload]
        )
    )

    const deleteUpload = useCallback(async (): Promise<void> => {
        if (!uploadOrError || isErrorLike(uploadOrError)) {
            return
        }

        let description = `${uploadOrError.inputCommit.slice(0, 7)}`
        if (uploadOrError.inputRoot) {
            description += ` rooted at ${uploadOrError.inputRoot}`
        }

        if (!window.confirm(`Delete upload for commit ${description}?`)) {
            return
        }

        setDeletionOrError('loading')

        try {
            await deleteLsifUpload({ id }).toPromise()
            setDeletionOrError('deleted')
        } catch (error) {
            setDeletionOrError(error)
        }
    }, [id, uploadOrError, deleteLsifUpload])

    const queryDependencies = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (uploadOrError && !isErrorLike(uploadOrError)) {
                return fetchLsifUploads({ dependencyOf: uploadOrError.id, ...args })
            }

            throw new Error('unreachable: queryDependencies referenced with invalid upload')
        },
        [uploadOrError, fetchLsifUploads]
    )

    const queryDependents = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (uploadOrError && !isErrorLike(uploadOrError)) {
                return fetchLsifUploads({ dependentOf: uploadOrError.id, ...args })
            }

            throw new Error('unreachable: queryDependents referenced with invalid upload')
        },
        [uploadOrError, fetchLsifUploads]
    )

    return deletionOrError === 'deleted' ? (
        <Redirect to="." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF upload" error={deletionOrError} />
    ) : (
        <div className="site-admin-lsif-upload-page w-100">
            <PageTitle title="Precise code intelligence uploads" />
            {isErrorLike(uploadOrError) ? (
                <ErrorAlert prefix="Error loading LSIF upload" error={uploadOrError} />
            ) : !uploadOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    <PageHeader
                        headingElement="h2"
                        path={[
                            {
                                text: `Upload for commit ${uploadOrError.projectRoot?.repository.name || ''}@${
                                    uploadOrError.projectRoot
                                        ? uploadOrError.projectRoot.commit.abbreviatedOID
                                        : uploadOrError.inputCommit.slice(0, 7)
                                }`,
                            },
                        ]}
                        className="mb-3"
                    />

                    <Container>
                        <CodeIntelUploadMeta node={uploadOrError} now={now} />
                    </Container>

                    <Container className="mt-2">
                        <CodeIntelStateBanner
                            state={uploadOrError.state}
                            placeInQueue={uploadOrError.placeInQueue}
                            failure={uploadOrError.failure}
                            typeName="upload"
                            pluralTypeName="uploads"
                            className={classNamesByState.get(uploadOrError.state)}
                        />
                        {uploadOrError.isLatestForRepo && (
                            <div>
                                <InformationOutlineIcon className="icon-inline" /> This upload can answer queries for
                                the tip of the default branch and are targets of cross-repository find reference
                                operations.
                            </div>
                        )}
                    </Container>

                    <Container className="mt-2">
                        <CodeIntelDeleteUpload
                            state={uploadOrError.state}
                            deleteUpload={deleteUpload}
                            deletionOrError={deletionOrError}
                        />
                    </Container>

                    <Container className="mt-2">
                        <CodeIntelAssociatedIndex node={uploadOrError} now={now} />
                        <h3>Timeline</h3>
                        <CodeIntelUploadTimeline now={now} upload={uploadOrError} className="mb-3" />
                    </Container>

                    {(uploadOrError.state === LSIFUploadState.COMPLETED ||
                        uploadOrError.state === LSIFUploadState.DELETING) && (
                        <Container className="mt-2">
                            <div className="mb-2">
                                {dependencyGraphState === DependencyGraphState.ShowDependencies ? (
                                    <h3>
                                        Dependencies
                                        <button
                                            type="button"
                                            className="btn btn-link float-right p-0 mb-2"
                                            onClick={() => setDependencyGraphState(DependencyGraphState.ShowDependents)}
                                        >
                                            Show dependents
                                        </button>
                                    </h3>
                                ) : (
                                    <h3>
                                        Dependents
                                        <button
                                            type="button"
                                            className="btn btn-link float-right p-0 mb-2"
                                            onClick={() =>
                                                setDependencyGraphState(DependencyGraphState.ShowDependencies)
                                            }
                                        >
                                            Show dependencies
                                        </button>
                                    </h3>
                                )}
                            </div>

                            {dependencyGraphState === DependencyGraphState.ShowDependencies ? (
                                <FilteredConnection
                                    listComponent="div"
                                    listClassName="codeintel-uploads__grid mb-3"
                                    noun="dependency"
                                    pluralNoun="dependencies"
                                    nodeComponent={DependencyOrDependentNode}
                                    queryConnection={queryDependencies}
                                    history={props.history}
                                    location={props.location}
                                    cursorPaging={true}
                                    emptyElement={<EmptyDependenciesElement />}
                                />
                            ) : (
                                <FilteredConnection
                                    listComponent="div"
                                    listClassName="codeintel-uploads__grid mb-3"
                                    noun="dependent"
                                    pluralNoun="dependents"
                                    nodeComponent={DependencyOrDependentNode}
                                    queryConnection={queryDependents}
                                    history={props.history}
                                    location={props.location}
                                    cursorPaging={true}
                                    emptyElement={<EmptyDependentsElement />}
                                />
                            )}
                        </Container>
                    )}
                </>
            )}
        </div>
    )
}

const terminalStates = new Set([LSIFUploadState.COMPLETED, LSIFUploadState.ERRORED, LSIFUploadState.DELETING])

function shouldReload(upload: LsifUploadFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(upload) && !(upload && terminalStates.has(upload.state))
}

interface CodeIntelUploadMetaProps {
    node: LsifUploadFields
    now?: () => Date
}

const CodeIntelUploadMeta: FunctionComponent<CodeIntelUploadMetaProps> = ({ node, now }) => (
    <div className="card">
        <div className="card-body">
            <div className="card border-0">
                <div className="card-body">
                    <h3 className="card-title">
                        <CodeIntelUploadOrIndexRepository node={node} />
                    </h3>

                    <p className="card-subtitle mb-2 text-muted">
                        <CodeIntelUploadOrIndexLastActivity node={{ ...node, queuedAt: null }} now={now} />
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

interface CodeIntelUploadTimelineProps {
    upload: LsifUploadFields
    now?: () => Date
    className?: string
}

enum FailedStage {
    UPLOADING,
    PROCESSING,
}

const CodeIntelUploadTimeline: FunctionComponent<CodeIntelUploadTimelineProps> = ({ upload, now, className }) => {
    let failedStage: FailedStage | null = null
    if (upload.state === LSIFUploadState.ERRORED && upload.startedAt === null) {
        failedStage = FailedStage.UPLOADING
    } else if (upload.state === LSIFUploadState.ERRORED && upload.startedAt !== null) {
        failedStage = FailedStage.PROCESSING
    }

    const stages = useMemo(
        () =>
            [uploadStages, processingStages, terminalStages].flatMap(stageConstructor =>
                stageConstructor(upload, failedStage)
            ),
        [upload, failedStage]
    )

    return <Timeline stages={stages} now={now} className={className} />
}

const uploadStages = (upload: LsifUploadFields, failedStage: FailedStage | null): TimelineStage[] => [
    {
        icon: <FileUploadIcon />,
        text:
            upload.state === LSIFUploadState.UPLOADING ||
            (LSIFUploadState.ERRORED && failedStage === FailedStage.UPLOADING)
                ? 'Upload started'
                : 'Uploaded',
        date: upload.uploadedAt,
        className:
            upload.state === LSIFUploadState.UPLOADING
                ? 'bg-primary'
                : upload.state === LSIFUploadState.ERRORED
                ? failedStage === FailedStage.UPLOADING
                    ? 'bg-danger'
                    : 'bg-success'
                : 'bg-success',
    },
]

const processingStages = (upload: LsifUploadFields, failedStage: FailedStage | null): TimelineStage[] => [
    {
        icon: <ProgressClockIcon />,
        text:
            upload.state === LSIFUploadState.PROCESSING ||
            (LSIFUploadState.ERRORED && failedStage === FailedStage.PROCESSING)
                ? 'Processing started'
                : 'Processed',
        date: upload.startedAt,
        className:
            upload.state === LSIFUploadState.PROCESSING
                ? 'bg-primary'
                : upload.state === LSIFUploadState.ERRORED
                ? 'bg-danger'
                : 'bg-success',
    },
]

const terminalStages = (upload: LsifUploadFields): TimelineStage[] =>
    upload.state === LSIFUploadState.COMPLETED
        ? [
              {
                  icon: <CheckIcon />,
                  text: 'Finished',
                  date: upload.finishedAt,
                  className: 'bg-success',
              },
          ]
        : upload.state === LSIFUploadState.ERRORED
        ? [
              {
                  icon: <ErrorIcon />,
                  text: 'Failed',
                  date: upload.finishedAt,
                  className: 'bg-danger',
              },
          ]
        : []

const CodeIntelAssociatedIndex: FunctionComponent<CodeIntelAssociatedIndexProps> = ({ node, now }) =>
    node.associatedIndex && node.projectRoot ? (
        <>
            <div className="list-group position-relative">
                <div className="codeintel-associated-index__grid mb-3">
                    <div className="d-flex flex-column codeintel-associated-index__information">
                        <div className="m-0">
                            <h3 className="m-0 d-block d-md-inline">This upload was created by an auto-indexing job</h3>
                        </div>

                        <div>
                            <small className="text-mute">
                                <CodeIntelUploadOrIndexLastActivity
                                    node={{ ...node.associatedIndex, uploadedAt: null }}
                                    now={now}
                                />
                            </small>
                        </div>
                    </div>

                    <span className="d-none d-md-inline codeintel-associated-index__state">
                        <CodeIntelState node={node.associatedIndex} className="d-flex flex-column align-items-center" />
                    </span>
                    <span>
                        <Link
                            to={`/${node.projectRoot.repository.name}/-/settings/code-intelligence/indexes/${node.associatedIndex.id}`}
                        >
                            <ChevronRightIcon />
                        </Link>
                    </span>

                    <span className="codeintel-associated-index__separator" />
                </div>
            </div>
        </>
    ) : (
        <></>
    )

interface DependencyOrDependentNodeProps {
    node: LsifUploadFields
    now?: () => Date
}

const DependencyOrDependentNode: FunctionComponent<DependencyOrDependentNodeProps> = ({ node }) => (
    <>
        <span className="codeintel-dependency-or-dependent-node__separator" />

        <div className="d-flex flex-column codeintel-dependency-or-dependent-node__information">
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
            </div>
        </div>

        <span className="d-none d-md-inline codeintel-dependency-or-dependent-node__state">
            <CodeIntelState node={node} className="d-flex flex-column align-items-center" />
        </span>
        <span>
            <Link to={`./${node.id}`}>
                <ChevronRightIcon />
            </Link>
        </span>
    </>
)

const EmptyDependenciesElement: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        This upload has no dependencies.
    </p>
)

const EmptyDependentsElement: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        This upload has no dependents.
    </p>
)

interface CodeIntelDeleteUploadProps {
    state: LSIFUploadState
    deleteUpload: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

const CodeIntelDeleteUpload: FunctionComponent<CodeIntelDeleteUploadProps> = ({
    state,
    deleteUpload,
    deletionOrError,
}) =>
    state === LSIFUploadState.DELETING ? (
        <></>
    ) : (
        <button
            type="button"
            className="btn btn-outline-danger"
            onClick={deleteUpload}
            disabled={deletionOrError === 'loading'}
            aria-describedby="upload-delete-button-help"
            data-tooltip={
                state === LSIFUploadState.COMPLETED
                    ? 'Deleting this upload will make it unavailable to answer code intelligence queries the next time the repository commit graph is refreshed.'
                    : 'Delete this upload immediately'
            }
        >
            <DeleteIcon className="icon-inline" /> Delete upload
        </button>
    )

interface CodeIntelAssociatedIndexProps {
    node: LsifUploadFields
    now?: () => Date
}
