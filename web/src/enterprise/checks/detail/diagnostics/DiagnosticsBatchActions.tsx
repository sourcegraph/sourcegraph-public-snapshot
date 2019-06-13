import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertIcon from 'mdi-react/AlertIcon'
import AnimationPlayIcon from 'mdi-react/AnimationPlayIcon'
import CheckAllIcon from 'mdi-react/CheckAllIcon'
import CheckBoxMultipleOutlineIcon from 'mdi-react/CheckBoxMultipleOutlineIcon'
import PlayCircleIcon from 'mdi-react/PlayCircleIcon'
import React, { useEffect, useState } from 'react'
import { from, Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { ActionsIcon } from '../../../../util/octicons'
import { CheckAreaContext } from '../CheckArea'
import { ChangesetPlanProps } from './useChangesetPlan'

interface Props
    extends Pick<ChangesetPlanProps, 'onChangesetPlanBatchActionClick'>,
        Pick<CheckAreaContext, 'checkProvider'>,
        ExtensionsControllerProps {
    parsedQuery: sourcegraph.DiagnosticQuery
    disabled?: boolean
    className?: string
}

const LOADING = 'loading' as const

/**
 * Buttons for performing batch actions on diagnostics.
 */
export const DiagnosticsBatchActions: React.FunctionComponent<Props> = ({
    parsedQuery,
    onChangesetPlanBatchActionClick,
    checkProvider,
    disabled,
    className = '',
}) => {
    const [batchActionsOrError, setBatchActionsOrError] = useState<
        typeof LOADING | sourcegraph.Operation[] | ErrorLike
    >(LOADING)
    const jsonParsedQuery = JSON.stringify(parsedQuery)
    useEffect(() => {
        const parsedQuery = JSON.parse(jsonParsedQuery) // avoid rerenders when object is not reference-equal
        const subscriptions = new Subscription()
        subscriptions.add(
            from(checkProvider.provideDiagnosticBatchActions(parsedQuery))
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setBatchActionsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [checkProvider, jsonParsedQuery])

    // TODO!(sqs): actually compute what the actions are instead of hardcoding

    return (
        <div className={`d-flex align-items-center w-100 ${className}`}>
            <div className="text-nowrap text-muted mr-3 py-2">
                <CheckBoxMultipleOutlineIcon className="icon-inline d-none" />
                <AnimationPlayIcon className="icon-inline" /> Batch actions:
            </div>
            {batchActionsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(batchActionsOrError) ? (
                <span title={batchActionsOrError.message}>
                    <AlertIcon className="icon-inline text-danger" />
                </span>
            ) : batchActionsOrError.length === 0 ? (
                <span className="text-muted">None</span>
            ) : (
                <div className="flex-1 d-flex overflow-auto">
                    {batchActionsOrError.map((op, i) => (
                        <button
                            key={i}
                            className={`btn ${i === 0 ? 'btn-primary' : 'btn-secondary'} text-nowrap py-1 mr-3 `}
                            disabled={disabled}
                            // tslint:disable-next-line: jsx-no-lambda
                            onClick={() => onChangesetPlanBatchActionClick(op)}
                        >
                            {op.message}
                        </button>
                    ))}
                </div>
            )}
        </div>
    )
}
