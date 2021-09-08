import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
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
import { Button, Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { LsifUploadFields } from '../../../graphql-operations'
import { fetchLsifUploads as defaultFetchLsifUploads } from '../shared/backend'
import { CodeIntelStateBanner } from '../shared/CodeIntelStateBanner'

import { deleteLsifUpload as defaultDeleteLsifUpload, fetchLsifUpload as defaultFetchUpload } from './backend'
import { CodeIntelAssociatedIndex } from './CodeIntelAssociatedIndex'
import { CodeIntelDeleteUpload } from './CodeIntelDeleteUpload'
import { CodeIntelUploadMeta } from './CodeIntelUploadMeta'
import styles from './CodeIntelUploadPage.module.scss'
import { CodeIntelUploadTimeline } from './CodeIntelUploadTimeline'
import { DependencyOrDependentNode } from './DependencyOrDependentNode'
import { EmptyDependencies } from './EmptyDependencies'
import { EmptyDependents } from './EmptyDependents'

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
                                        <Button
                                            type="button"
                                            className="float-right p-0 mb-2"
                                            variant="link"
                                            onClick={() => setDependencyGraphState(DependencyGraphState.ShowDependents)}
                                        >
                                            Show dependents
                                        </Button>
                                    </h3>
                                ) : (
                                    <h3>
                                        Dependents
                                        <Button
                                            type="button"
                                            className="float-right p-0 mb-2"
                                            variant="link"
                                            onClick={() =>
                                                setDependencyGraphState(DependencyGraphState.ShowDependencies)
                                            }
                                        >
                                            Show dependencies
                                        </Button>
                                    </h3>
                                )}
                            </div>

                            {dependencyGraphState === DependencyGraphState.ShowDependencies ? (
                                <FilteredConnection
                                    listComponent="div"
                                    listClassName={classNames(styles.grid, 'mb-3')}
                                    noun="dependency"
                                    pluralNoun="dependencies"
                                    nodeComponent={DependencyOrDependentNode}
                                    queryConnection={queryDependencies}
                                    history={props.history}
                                    location={props.location}
                                    cursorPaging={true}
                                    emptyElement={<EmptyDependencies />}
                                />
                            ) : (
                                <FilteredConnection
                                    listComponent="div"
                                    listClassName={classNames(styles.grid, 'mb-3')}
                                    noun="dependent"
                                    pluralNoun="dependents"
                                    nodeComponent={DependencyOrDependentNode}
                                    queryConnection={queryDependents}
                                    history={props.history}
                                    location={props.location}
                                    cursorPaging={true}
                                    emptyElement={<EmptyDependents />}
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
