import * as H from 'history'
import React, { useCallback, useState } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isErrorLike, asError } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../components/alerts'
import { Scalars } from '../../../graphql-operations'

import { closeBatchChange as _closeBatchChange } from './backend'

export interface BatchChangeCloseAlertProps {
    batchChangeID: Scalars['ID']
    batchChangeURL: string
    closeChangesets: boolean
    viewerCanAdminister: boolean
    totalCount: number
    setCloseChangesets: (newValue: boolean) => void
    history: H.History

    /** For testing only. */
    closeBatchChange?: typeof _closeBatchChange
}

export const BatchChangeCloseAlert: React.FunctionComponent<BatchChangeCloseAlertProps> = ({
    batchChangeID,
    batchChangeURL,
    closeChangesets,
    totalCount,
    setCloseChangesets,
    viewerCanAdminister,
    history,
    closeBatchChange = _closeBatchChange,
}) => {
    const onChangeCloseChangesets = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            setCloseChangesets(event.target.checked)
        },
        [setCloseChangesets]
    )
    const onCancel = useCallback<React.MouseEventHandler>(() => {
        history.push(batchChangeURL)
    }, [history, batchChangeURL])
    const [isClosing, setIsClosing] = useState<boolean | Error>(false)
    const onClose = useCallback<React.MouseEventHandler>(async () => {
        setIsClosing(true)
        try {
            await closeBatchChange({ batchChange: batchChangeID, closeChangesets })
            history.push(batchChangeURL)
        } catch (error) {
            setIsClosing(asError(error))
        }
    }, [history, closeChangesets, closeBatchChange, batchChangeID, batchChangeURL])
    return (
        <>
            <div className="card mb-3">
                <div className="card-body p-3">
                    <p>
                        <strong>
                            After closing this batch change, it will be read-only and no new batch specs can be applied.
                        </strong>
                    </p>
                    {totalCount > 0 && (
                        <>
                            <p>By default, all changesets remain untouched.</p>
                            <div className="form-check mb-3">
                                <input
                                    id="closeChangesets"
                                    type="checkbox"
                                    checked={closeChangesets}
                                    onChange={onChangeCloseChangesets}
                                    className="test-batches-close-changesets-checkbox form-check-input"
                                    disabled={isClosing === true || !viewerCanAdminister}
                                />
                                <label className="form-check-label" htmlFor="closeChangesets">
                                    Also close all {totalCount} open changesets on code hosts.
                                </label>
                            </div>
                        </>
                    )}
                    {!viewerCanAdminister && (
                        <div className="alert alert-warning">
                            You don't have permission to close this batch change. See{' '}
                            <a
                                className="alert-link"
                                href="https://docs.sourcegraph.com/batch_changes/explanations/permissions_in_batch_changes"
                            >
                                Permissions in batch changes
                            </a>{' '}
                            for more information about the batch changes permission model.
                        </div>
                    )}
                    <div className="d-flex justify-content-end">
                        <button
                            type="button"
                            className="btn btn-secondary mr-2 test-batches-close-abort-btn"
                            onClick={onCancel}
                            disabled={isClosing === true || !viewerCanAdminister}
                        >
                            Cancel
                        </button>
                        <button
                            type="button"
                            className="btn btn-danger test-batches-confirm-close-btn"
                            onClick={onClose}
                            disabled={isClosing === true || !viewerCanAdminister}
                        >
                            {isClosing === true && <LoadingSpinner className="icon-inline" />} Close batch change
                        </button>
                    </div>
                </div>
            </div>
            {isErrorLike(isClosing) && <ErrorAlert error={isClosing} />}
        </>
    )
}
