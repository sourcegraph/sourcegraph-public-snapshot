import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
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
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { LsifUploadFields } from '../../../graphql-operations'
import { fetchLsifUploads as defaultFetchLsifUploads } from '../list/backend'
import { CodeIntelState } from '../shared/CodeIntelState'
import { CodeIntelStateBanner } from '../shared/CodeIntelStateBanner'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexRoot } from '../shared/CodeIntelUploadOrIndexRoot'

import { deleteLsifUpload, fetchLsifUpload as defaultFetchUpload } from './backend'
import { CodeIntelAssociatedIndex } from './CodeIntelAssociatedIndex'
import { CodeIntelUploadMeta } from './CodeIntelUploadMeta'
import { CodeIntelUploadTimeline } from './CodeIntelUploadTimeline'

export interface CodeIntelUploadPageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {
    fetchLsifUpload?: typeof defaultFetchUpload
    fetchLsifUploads?: typeof defaultFetchLsifUploads
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
    }, [id, uploadOrError])

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

    const showDependencyGraphState = useCallback(
        (state: DependencyGraphState) => () => setDependencyGraphState(state),
        [setDependencyGraphState]
    )

    return deletionOrError === 'deleted' ? (
        <Redirect to="." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF upload" error={deletionOrError} />
    ) : (
        <div className="site-admin-lsif-upload-page w-100">
            <PageTitle title="Code intelligence - uploads" />
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
                                text: `Upload for commit ${
                                    uploadOrError.projectRoot
                                        ? uploadOrError.projectRoot.commit.abbreviatedOID
                                        : uploadOrError.inputCommit.slice(0, 7)
                                } indexed by ${uploadOrError.inputIndexer} rooted at ${
                                    (uploadOrError.projectRoot
                                        ? uploadOrError.projectRoot.path
                                        : uploadOrError.inputRoot) || '/'
                                }`,
                            },
                        ]}
                        className="mb-3"
                    />

                    <Container>
                        <CodeIntelStateBanner
                            state={uploadOrError.state}
                            placeInQueue={uploadOrError.placeInQueue}
                            failure={uploadOrError.failure}
                            typeName="upload"
                            pluralTypeName="uploads"
                            className={classNamesByState.get(uploadOrError.state)}
                        />
                        {uploadOrError.isLatestForRepo && (
                            <div className="mb-3">
                                <InformationOutlineIcon className="icon-inline" /> This upload can answer queries for
                                the tip of the default branch and are targets of cross-repository find reference
                                operations.
                            </div>
                        )}
                        <CodeIntelUploadMeta node={uploadOrError} now={now} />
                        <CodeIntelAssociatedIndex node={uploadOrError} now={now} />

                        <h3>Timeline</h3>
                        <CodeIntelUploadTimeline now={now} upload={uploadOrError} className="mb-3" />
                    </Container>

                    <Container className="mt-2">
                        <div className="mb-2">
                            {dependencyGraphState === DependencyGraphState.ShowDependencies ? (
                                <h2>
                                    Dependencies
                                    <button
                                        type="button"
                                        className="btn btn-link float-right p-0 mb-2"
                                        onClick={showDependencyGraphState(DependencyGraphState.ShowDependents)}
                                    >
                                        Show dependents
                                    </button>
                                </h2>
                            ) : (
                                <h2>
                                    Dependents
                                    <button
                                        type="button"
                                        className="btn btn-link float-right p-0 mb-2"
                                        onClick={showDependencyGraphState(DependencyGraphState.ShowDependencies)}
                                    >
                                        Show dependencies
                                    </button>
                                </h2>
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

                    <Container className="mt-2">
                        <CodeIntelDeleteUpload
                            state={uploadOrError.state}
                            deleteUpload={deleteUpload}
                            deletionOrError={deletionOrError}
                        />
                    </Container>
                </>
            )}
        </div>
    )
}

const terminalStates = new Set([LSIFUploadState.COMPLETED, LSIFUploadState.ERRORED])

function shouldReload(upload: LsifUploadFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(upload) && !(upload && terminalStates.has(upload.state))
}

interface DependencyOrDependentNodeProps {
    node: LsifUploadFields
    now?: () => Date
}

const DependencyOrDependentNode: FunctionComponent<DependencyOrDependentNodeProps> = ({ node }) => (
    <>
        <span className="codeintel-upload-node__separator" />

        <div className="d-flex flex-column codeintel-upload-node__information">
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

        <span className="d-none d-md-inline codeintel-upload-node__state">
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
