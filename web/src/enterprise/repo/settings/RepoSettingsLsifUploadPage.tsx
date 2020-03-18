import * as GQL from '../../../../../shared/src/graphql/schema'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ClockOutlineIcon from 'mdi-react/ClockOutlineIcon'
import React, { FunctionComponent, useEffect, useMemo, useState } from 'react'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { catchError, takeWhile, concatMap } from 'rxjs/operators'
import { ErrorAlert } from '../../../components/alerts'
import { eventLogger } from '../../../tracking/eventLogger'
import { fetchLsifUpload, deleteLsifUpload } from './backend'
import { Link } from '../../../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../../../components/PageTitle'
import { RouteComponentProps, Redirect } from 'react-router'
import { Timestamp } from '../../../components/time/Timestamp'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import DeleteIcon from 'mdi-react/DeleteIcon'
import { SchedulerLike, timer } from 'rxjs'

const REFRESH_INTERVAL_MS = 5000

interface Props extends RouteComponentProps<{ id: string }> {
    repo: GQL.IRepository

    /** Scheduler for the refresh timer */
    scheduler?: SchedulerLike
}

const terminalStates = [GQL.LSIFUploadState.COMPLETED, GQL.LSIFUploadState.ERRORED]

function shouldReload(v: GQL.ILSIFUpload | ErrorLike | null | undefined): boolean {
    return !isErrorLike(v) && !(v && terminalStates.includes(v.state))
}

/**
 * A page displaying metadata about an LSIF upload.
 */
export const RepoSettingsLsifUploadPage: FunctionComponent<Props> = ({
    repo,
    scheduler,
    match: {
        params: { id },
    },
}) => {
    useEffect(() => eventLogger.logViewEvent('RepoSettingsLsifUpload'))

    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()

    const uploadOrError = useObservable(
        useMemo(
            () =>
                timer(0, REFRESH_INTERVAL_MS, scheduler).pipe(
                    concatMap(() => fetchLsifUpload({ id }).pipe(catchError((error): [ErrorLike] => [asError(error)]))),
                    takeWhile(shouldReload, true)
                ),
            [id, scheduler]
        )
    )

    const deleteUpload = async (): Promise<void> => {
        if (!uploadOrError || isErrorLike(uploadOrError)) {
            return
        }

        let description = `commit ${uploadOrError.inputCommit.substring(0, 7)}`
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
        } catch (err) {
            setDeletionOrError(err)
        }
    }

    return deletionOrError === 'deleted' ? (
        <Redirect to=".." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF upload" error={deletionOrError} />
    ) : (
        <div className="site-admin-lsif-upload-page w-100">
            <PageTitle title="LSIF uploads - Admin" />
            {isErrorLike(uploadOrError) ? (
                <ErrorAlert prefix="Error loading LSIF upload" error={uploadOrError} />
            ) : !uploadOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    <div className="mb-1">
                        <h2 className="mb-0">
                            Upload for commit{' '}
                            {uploadOrError.projectRoot
                                ? uploadOrError.projectRoot.commit.abbreviatedOID
                                : uploadOrError.inputCommit.substring(0, 7)}{' '}
                            indexed by {uploadOrError.inputIndexer} rooted at{' '}
                            {uploadOrError.projectRoot?.path || uploadOrError.inputRoot || '/'}
                        </h2>
                    </div>

                    {uploadOrError.state === GQL.LSIFUploadState.PROCESSING ? (
                        <div className="alert alert-primary mb-4 mt-3">
                            <LoadingSpinner className="icon-inline" />{' '}
                            <span className="e2e-upload-state">Upload is currently being processed...</span>
                        </div>
                    ) : uploadOrError.state === GQL.LSIFUploadState.COMPLETED ? (
                        <div className="alert alert-success mb-4 mt-3">
                            <CheckIcon className="icon-inline" />{' '}
                            <span className="e2e-upload-state">Upload processed successfully.</span>
                        </div>
                    ) : uploadOrError.state === GQL.LSIFUploadState.ERRORED ? (
                        <div className="alert alert-danger mb-4 mt-3">
                            <AlertCircleIcon className="icon-inline" />{' '}
                            <span className="e2e-upload-state">Upload failed to complete:</span>{' '}
                            <code>{uploadOrError.failure && uploadOrError.failure.summary}</code>
                        </div>
                    ) : (
                        <div className="alert alert-primary mb-4 mt-3">
                            <ClockOutlineIcon className="icon-inline" />{' '}
                            <span className="e2e-upload-state">
                                Upload is queued. There are {uploadOrError.placeInQueue} uploads ahead of this one.
                            </span>
                        </div>
                    )}

                    <table className="table">
                        <tbody>
                            <tr>
                                <td>Repository</td>
                                <td>
                                    {uploadOrError.projectRoot ? (
                                        <Link to={uploadOrError.projectRoot.commit.repository.url}>
                                            {uploadOrError.projectRoot.commit.repository.name}
                                        </Link>
                                    ) : (
                                        repo.name
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>Commit</td>
                                <td>
                                    {uploadOrError.projectRoot ? (
                                        <Link to={uploadOrError.projectRoot.commit.url}>
                                            <code>{uploadOrError.projectRoot.commit.oid}</code>
                                        </Link>
                                    ) : (
                                        uploadOrError.inputCommit
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>Root</td>
                                <td>
                                    {uploadOrError.projectRoot ? (
                                        <Link to={uploadOrError.projectRoot.url}>
                                            <strong>{uploadOrError.projectRoot.path || '/'}</strong>
                                        </Link>
                                    ) : (
                                        uploadOrError.inputRoot || '/'
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>Indexer</td>
                                <td>{uploadOrError.inputIndexer}</td>
                            </tr>

                            <tr>
                                <td>Is latest for repo</td>
                                <td>
                                    {uploadOrError.finishedAt ? (
                                        <span className="e2e-is-latest-for-repo">
                                            {uploadOrError.isLatestForRepo ? 'yes' : 'no'}
                                        </span>
                                    ) : (
                                        <span className="text-muted">Upload has not yet completed.</span>
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>Uploaded</td>
                                <td>
                                    <Timestamp date={uploadOrError.uploadedAt} noAbout={true} />
                                </td>
                            </tr>

                            <tr>
                                <td>Began processing</td>
                                <td>
                                    {uploadOrError.startedAt ? (
                                        <Timestamp date={uploadOrError.startedAt} noAbout={true} />
                                    ) : (
                                        <span className="text-muted">Upload has not yet started.</span>
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>
                                    {uploadOrError.state === GQL.LSIFUploadState.ERRORED && uploadOrError.finishedAt
                                        ? 'Failed'
                                        : 'Finished'}{' '}
                                    processing
                                </td>
                                <td>
                                    {uploadOrError.finishedAt ? (
                                        <Timestamp date={uploadOrError.finishedAt} noAbout={true} />
                                    ) : (
                                        <span className="text-muted">Upload has not yet completed.</span>
                                    )}
                                </td>
                            </tr>
                        </tbody>
                    </table>

                    <div className="action-container">
                        <div className="action-container__row">
                            <div className="action-container__description">
                                <h4 className="action-container__title">Delete this upload</h4>
                                <div>
                                    Deleting this upload make it immediately unavailable to answer code intelligence
                                    queries.
                                </div>
                            </div>
                            <div className="action-container__btn-container">
                                <button
                                    type="button"
                                    className="btn btn-danger action-container__btn"
                                    onClick={deleteUpload}
                                    disabled={deletionOrError === 'loading'}
                                    data-tooltip="Delete upload"
                                >
                                    <DeleteIcon className="icon-inline" />
                                </button>
                            </div>
                        </div>
                    </div>
                </>
            )}
        </div>
    )
}
