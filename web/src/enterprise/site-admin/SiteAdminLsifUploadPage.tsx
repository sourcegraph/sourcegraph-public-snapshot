import * as GQL from '../../../../shared/src/graphql/schema'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ClockOutlineIcon from 'mdi-react/ClockOutlineIcon'
import React, { FunctionComponent, useEffect, useMemo } from 'react'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { catchError } from 'rxjs/operators'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchLsifUpload } from './backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Timestamp } from '../../components/time/Timestamp'
import { useObservable } from '../../util/useObservable'
import { ErrorAlert } from '../../components/alerts'
import { Link } from '../../../../shared/src/components/Link'

interface Props extends RouteComponentProps<{ id: string }> {}

/**
 * A page displaying metadata about an LSIF upload.
 */
export const SiteAdminLsifUploadPage: FunctionComponent<Props> = ({
    match: {
        params: { id },
    },
}) => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminLsifUpload'))

    const uploadOrError = useObservable(
        useMemo(() => fetchLsifUpload({ id }).pipe(catchError((error): [ErrorLike] => [asError(error)])), [id])
    )

    return (
        <div className="site-admin-lsif-upload-page w-100">
            <PageTitle title="LSIF uploads - Admin" />
            {!uploadOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(uploadOrError) ? (
                <div className="alert alert-danger">
                    <ErrorAlert prefix="Error loading LSIF upload" error={uploadOrError} />
                </div>
            ) : (
                <>
                    <div className="mt-3 mb-1">
                        <h2 className="mb-0">
                            Upload for{' '}
                            {uploadOrError.projectRoot
                                ? lsifUploadDescription(
                                      uploadOrError.projectRoot.commit.repository.name,
                                      uploadOrError.projectRoot.commit.abbreviatedOID,
                                      uploadOrError.projectRoot.path
                                  )
                                : lsifUploadDescription(
                                      uploadOrError.inputRepoName,
                                      uploadOrError.inputCommit.substring(0, 7),
                                      uploadOrError.inputRoot
                                  )}
                        </h2>
                    </div>

                    {uploadOrError.state === GQL.LSIFUploadState.PROCESSING ? (
                        <div className="alert alert-primary mb-4 mt-3">
                            <LoadingSpinner className="icon-inline" /> Upload is currently being processed...
                        </div>
                    ) : uploadOrError.state === GQL.LSIFUploadState.COMPLETED ? (
                        <div className="alert alert-success mb-4 mt-3">
                            <CheckIcon className="icon-inline" /> Upload completed successfully.
                        </div>
                    ) : uploadOrError.state === GQL.LSIFUploadState.ERRORED ? (
                        <div className="alert alert-danger mb-4 mt-3">
                            <AlertCircleIcon className="icon-inline" /> Upload failed to complete:{' '}
                            <code>{uploadOrError.failure && uploadOrError.failure.summary}</code>
                        </div>
                    ) : (
                        <div className="alert alert-primary mb-4 mt-3">
                            <ClockOutlineIcon className="icon-inline" /> Upload is queued.
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
                                        uploadOrError.inputRepoName
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
                                <td>Queued</td>
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
                </>
            )}
        </div>
    )
}

export function lsifUploadDescription(repoName: string, commit: string, root: string): string {
    return `${repoName}@${commit}${root === '' ? '' : ` rooted at ${root}`}`
}
