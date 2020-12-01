import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import classNames from 'classnames'
import * as H from 'history'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { FunctionComponent, ReactNode, useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { SchedulerLike, timer } from 'rxjs'
import { catchError, concatMap, delay, repeatWhen, takeWhile } from 'rxjs/operators'
import { LSIFIndexState } from '../../../../../shared/src/graphql-operations'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { ErrorAlert } from '../../../components/alerts'
import { PageHeader } from '../../../components/PageHeader'
import { PageTitle } from '../../../components/PageTitle'
import { LsifIndexFields } from '../../../graphql-operations'
import { CodeIntelStateBanner } from '../shared/CodeIntelStateBanner'
import { deleteLsifIndex, fetchLsifIndex as defaultFetchLsifIndex } from './backend'
import { CodeIntelIndexMeta } from './CodeIntelIndexMeta'
import { CodeIntelIndexTimeline } from './CodeIntelIndexTimeline'

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
        <div className="site-admin-lsif-index-page w-100 web-content">
            <PageTitle title="Code intelligence - auto-indexing" />
            {isErrorLike(indexOrError) ? (
                <ErrorAlert prefix="Error loading LSIF index" error={indexOrError} history={history} />
            ) : !indexOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    <CodeIntelIndexPageTitle
                        index={indexOrError}
                        actions={<CodeIntelDeleteIndex deleteIndex={deleteIndex} deletionOrError={deletionOrError} />}
                        className="mb-2"
                    />
                    <CodeIntelStateBanner
                        state={indexOrError.state}
                        placeInQueue={indexOrError.placeInQueue}
                        failure={indexOrError.failure}
                        typeName="index"
                        pluralTypeName="indexes"
                        history={history}
                        className={classNames('mb-3', classNamesByState.get(indexOrError.state))}
                    />
                    <div className="card mb-3">
                        <div className="card-body">
                            <CodeIntelIndexMeta node={indexOrError} now={now} />
                        </div>
                    </div>
                    <CodeIntelIndexTimeline index={indexOrError} now={now} className="mb-3" />
                </>
            )}
        </div>
    )
}

const terminalStates = new Set([LSIFIndexState.COMPLETED, LSIFIndexState.ERRORED])

function shouldReload(index: LsifIndexFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(index) && !(index && terminalStates.has(index.state))
}

interface CodeIntelIndexPageTitleProps {
    index: LsifIndexFields
    actions?: ReactNode
    className?: string
}

const CodeIntelIndexPageTitle: FunctionComponent<CodeIntelIndexPageTitleProps> = ({ index, actions, className }) => (
    <PageHeader
        title={
            <>
                <span className="text-muted">Auto-index record for commit</span>
                <span className="ml-2">
                    {index.projectRoot ? index.projectRoot.commit.abbreviatedOID : index.inputCommit.slice(0, 7)}
                </span>
            </>
        }
        actions={actions}
        className={className}
    />
)

interface CodeIntelDeleteIndexProps {
    deleteIndex: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

const CodeIntelDeleteIndex: FunctionComponent<CodeIntelDeleteIndexProps> = ({ deleteIndex, deletionOrError }) => (
    <button
        type="button"
        className="btn btn-outline-danger"
        onClick={deleteIndex}
        disabled={deletionOrError === 'loading'}
        aria-describedby="upload-delete-button-help"
        data-tooltip="Deleting this index will remove it from the index queue."
    >
        <DeleteIcon className="icon-inline" /> Delete index
    </button>
)
