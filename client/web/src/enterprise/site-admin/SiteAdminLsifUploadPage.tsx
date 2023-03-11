import { useEffect, useMemo } from 'react'

import { Navigate, useParams } from 'react-router-dom'
import { catchError } from 'rxjs/operators'

import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner, useObservable, ErrorAlert } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

import { fetchLsifUpload } from './backend'

/**
 * A page displaying metadata about an LSIF upload.
 */
export const SiteAdminLsifUploadPage: React.FC<{}> = () => {
    const { id = '' } = useParams<{ id: string }>()
    useEffect(() => eventLogger.logViewEvent('SiteAdminLsifUpload'))

    const uploadOrError = useObservable(
        useMemo(() => fetchLsifUpload({ id }).pipe(catchError((error): [ErrorLike] => [asError(error)])), [id])
    )

    return (
        <div className="site-admin-lsif-upload-page w-100">
            <PageTitle title="LSIF uploads - Admin" />
            {!uploadOrError ? (
                <LoadingSpinner />
            ) : isErrorLike(uploadOrError) ? (
                <ErrorAlert prefix="Error loading LSIF upload" error={uploadOrError} />
            ) : !uploadOrError.projectRoot ? (
                <ErrorAlert prefix="Error loading LSIF upload" error={{ message: 'Cannot resolve project root' }} />
            ) : (
                <Navigate
                    replace={true}
                    to={`${uploadOrError.projectRoot.commit.repository.url}/-/code-graph/uploads/${id}`}
                />
            )}
        </div>
    )
}
