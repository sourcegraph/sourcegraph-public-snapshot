import { useEffect, useMemo } from 'react'

import { Navigate, useParams } from 'react-router-dom'
import { catchError } from 'rxjs/operators'

import { asError, type ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { LoadingSpinner, useObservable, ErrorAlert } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

import { fetchPreciseIndex } from './backend'

/**
 * A page displaying metadata about a precise index.
 */
export const SiteAdminPreciseIndexPage: React.FC<TelemetryV2Props> = props => {
    const { id = '' } = useParams<{ id: string }>()
    useEffect(() => {
        props.telemetryRecorder.recordEvent('siteAdminPreciseIndex', 'viewed')
        eventLogger.logViewEvent('SiteAdminPreciseIndex')
    })

    const indexOrError = useObservable(
        useMemo(() => fetchPreciseIndex({ id }).pipe(catchError((error): [ErrorLike] => [asError(error)])), [id])
    )

    return (
        <div className="site-admin-lsif-upload-page w-100">
            <PageTitle title="Precise indexes - Admin" />
            {!indexOrError ? (
                <LoadingSpinner />
            ) : isErrorLike(indexOrError) ? (
                <ErrorAlert prefix="Error loading precise index" error={indexOrError} />
            ) : !indexOrError.projectRoot ? (
                <ErrorAlert prefix="Error loading precise index" error={{ message: 'Cannot resolve project root' }} />
            ) : (
                <Navigate
                    replace={true}
                    to={`${indexOrError.projectRoot.commit.repository.url}/-/code-graph/uploads/${id}`}
                />
            )}
        </div>
    )
}
