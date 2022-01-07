import Dialog from '@reach/dialog'
import React, { useCallback, useState } from 'react'

import { asError, isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../../components/alerts'
import { Scalars } from '../../../../graphql-operations'
import { closeChangesets as _closeChangesets } from '../backend'

export interface CloseChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]

    /** For testing only. */
    closeChangesets?: typeof _closeChangesets
}

export const CloseChangesetsModal: React.FunctionComponent<CloseChangesetsModalProps> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
    closeChangesets = _closeChangesets,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)

    const onSubmit = useCallback<React.FormEventHandler>(async () => {
        setIsLoading(true)
        try {
            await closeChangesets(batchChangeID, changesetIDs)
            afterCreate()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [changesetIDs, closeChangesets, batchChangeID, afterCreate])

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={MODAL_LABEL_ID}
        >
            <h3 id={MODAL_LABEL_ID}>Close changesets</h3>
            <p className="mb-4">Are you sure you want to close all the selected changesets on the code hosts?</p>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
            <div className="d-flex justify-content-end">
                <button
                    type="button"
                    disabled={isLoading === true}
                    className="btn btn-outline-secondary mr-2"
                    onClick={onCancel}
                >
                    Cancel
                </button>
                <button type="button" onClick={onSubmit} disabled={isLoading === true} className="btn btn-primary">
                    {isLoading === true && <LoadingSpinner />}
                    Close
                </button>
            </div>
        </Dialog>
    )
}

const MODAL_LABEL_ID = 'close-changesets-modal-title'
