import React, { FunctionComponent, useEffect, useMemo } from 'react'
import { RouteComponentProps, Redirect } from 'react-router'
import { catchError } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { ErrorAlert } from '../../components/alerts'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

import { fetchLsifUpload } from './backend'

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
                <ErrorAlert prefix="Error loading LSIF upload" error={uploadOrError} />
            ) : !uploadOrError.projectRoot ? (
                <ErrorAlert prefix="Error loading LSIF upload" error={{ message: 'Cannot resolve project root' }} />
            ) : (
                <Redirect
                    to={`${uploadOrError.projectRoot.repository.url}/-/settings/code-intelligence/lsif-uploads/${id}`}
                />
            )}
        </div>
    )
}
