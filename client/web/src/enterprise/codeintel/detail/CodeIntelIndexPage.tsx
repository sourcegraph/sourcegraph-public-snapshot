import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { SchedulerLike, timer } from 'rxjs'
import { catchError, concatMap, delay, repeatWhen, takeWhile } from 'rxjs/operators'
import { LSIFIndexState } from '../../../../../shared/src/graphql-operations'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { LsifIndexFields } from '../../../graphql-operations'
import { deleteLsifIndex, fetchLsifIndex as defaultFetchLsifIndex } from './backend'
import { CodeIntelDeleteIndex } from './CodeIntelDeleteIndex'
import { CodeIntelIndexInfo } from './CodeIntelIndexInfo'
import { CodeIntelIndexPageTitle } from './CodeIntelIndexPageTitle'
import { CodeIntelStateBanner } from './CodeIntelStateBanner'

export interface CodeIntelIndexPageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {
    fetchLsifIndex?: typeof defaultFetchLsifIndex
    now?: () => Date
    /** Scheduler for the refresh timer */
    scheduler?: SchedulerLike
    history: H.History
}

const REFRESH_INTERVAL_MS = 5000

const classNamesByState = new Map([
    [LSIFIndexState.COMPLETED, 'alert-success'],
    [LSIFIndexState.ERRORED, 'alert-danger'],
])

export const CodeIntelIndexPage: FunctionComponent<CodeIntelIndexPageProps> = ({
    scheduler,
    match: {
        params: { id },
    },
    history,
    telemetryService,
    fetchLsifIndex = defaultFetchLsifIndex,
    now,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndex'), [telemetryService])

    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()

    const indexOrError = useObservable(
        useMemo(
            () =>
                timer(0, REFRESH_INTERVAL_MS, scheduler).pipe(
                    concatMap(() =>
                        fetchLsifIndex({ id }).pipe(
                            catchError((error): [ErrorLike] => [asError(error)]),
                            repeatWhen(observable => observable.pipe(delay(REFRESH_INTERVAL_MS)))
                        )
                    ),
                    takeWhile(shouldReload, true)
                ),
            [id, scheduler, fetchLsifIndex]
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
    }, [id, indexOrError])

    return deletionOrError === 'deleted' ? (
        <Redirect to="." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF index record" error={deletionOrError} history={history} />
    ) : (
        <div className="site-admin-lsif-index-page w-100">
            <PageTitle title="Code intelligence - auto-indexing" />
            {isErrorLike(indexOrError) ? (
                <ErrorAlert prefix="Error loading LSIF index" error={indexOrError} history={history} />
            ) : !indexOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    <CodeIntelIndexPageTitle index={indexOrError} />
                    <CodeIntelStateBanner
                        state={indexOrError.state}
                        placeInQueue={indexOrError.placeInQueue}
                        failure={indexOrError.failure}
                        typeName="index"
                        pluralTypeName="indexes"
                        history={history}
                        className={classNamesByState.get(indexOrError.state)}
                    />
                    <CodeIntelIndexInfo index={indexOrError} now={now} />
                    <CodeIntelDeleteIndex deleteIndex={deleteIndex} deletionOrError={deletionOrError} />
                </>
            )}
        </div>
    )
}

const terminalStates = new Set([LSIFIndexState.COMPLETED, LSIFIndexState.ERRORED])

function shouldReload(index: LsifIndexFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(index) && !(index && terminalStates.has(index.state))
}
