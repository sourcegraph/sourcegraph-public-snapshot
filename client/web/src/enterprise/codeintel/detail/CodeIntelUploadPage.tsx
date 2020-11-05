import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { SchedulerLike, timer } from 'rxjs'
import { catchError, concatMap, delay, repeatWhen, takeWhile } from 'rxjs/operators'
import { LSIFUploadState } from '../../../../../shared/src/graphql-operations'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { LsifUploadFields } from '../../../graphql-operations'
import { deleteLsifUpload, fetchLsifUpload as defaultFetchUpload } from './backend'
import { CodeIntelDeleteUpload } from './CodeIntelDeleteUpload'
import { CodeIntelStateBanner } from './CodeIntelStateBanner'
import { CodeIntelUploadInfo } from './CodeIntelUploadInfo'
import { CodeIntelUploadPageTitle } from './CodeIntelUploadPageTitle'

export interface CodeIntelUploadPageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {
    fetchLsifUpload?: typeof defaultFetchUpload
    now?: () => Date
    /** Scheduler for the refresh timer */
    scheduler?: SchedulerLike
    history: H.History
}

const REFRESH_INTERVAL_MS = 5000

const classNamesByState = new Map([
    [LSIFUploadState.COMPLETED, 'alert-success'],
    [LSIFUploadState.ERRORED, 'alert-danger'],
])

export const CodeIntelUploadPage: FunctionComponent<CodeIntelUploadPageProps> = ({
    scheduler,
    match: {
        params: { id },
    },
    history,
    fetchLsifUpload = defaultFetchUpload,
    telemetryService,
    now,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelUpload'), [telemetryService])

    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()

    const uploadOrError = useObservable(
        useMemo(
            () =>
                timer(0, REFRESH_INTERVAL_MS, scheduler).pipe(
                    concatMap(() =>
                        fetchLsifUpload({ id }).pipe(
                            catchError((error): [ErrorLike] => [asError(error)]),
                            repeatWhen(observable => observable.pipe(delay(REFRESH_INTERVAL_MS)))
                        )
                    ),
                    takeWhile(shouldReload, true)
                ),
            [id, scheduler, fetchLsifUpload]
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

    return deletionOrError === 'deleted' ? (
        <Redirect to="." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF upload" error={deletionOrError} history={history} />
    ) : (
        <div className="site-admin-lsif-upload-page w-100">
            <PageTitle title="Code intelligence - uploads" />

            {isErrorLike(uploadOrError) ? (
                <ErrorAlert prefix="Error loading LSIF upload" error={uploadOrError} history={history} />
            ) : !uploadOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    <CodeIntelUploadPageTitle upload={uploadOrError} />
                    <CodeIntelStateBanner
                        state={uploadOrError.state}
                        placeInQueue={uploadOrError.placeInQueue}
                        failure={uploadOrError.failure}
                        typeName="upload"
                        pluralTypeName="uploads"
                        history={history}
                        className={classNamesByState.get(uploadOrError.state)}
                    />
                    <CodeIntelUploadInfo upload={uploadOrError} now={now} />
                    <CodeIntelDeleteUpload deleteUpload={deleteUpload} deletionOrError={deletionOrError} />
                </>
            )}
        </div>
    )
}

const terminalStates = new Set([LSIFUploadState.COMPLETED, LSIFUploadState.ERRORED])

function shouldReload(upload: LsifUploadFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(upload) && !(upload && terminalStates.has(upload.state))
}
