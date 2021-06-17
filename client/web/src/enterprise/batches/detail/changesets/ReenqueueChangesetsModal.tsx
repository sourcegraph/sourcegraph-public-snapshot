import Dialog from '@reach/dialog'
import React, { useCallback, useState } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../components/alerts'
import { Scalars } from '../../../../graphql-operations'
import { reenqueueChangesets as _reenqueueChangesets } from '../backend'

export interface ReenqueueChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: () => Promise<Scalars['ID'][]>

    /** For testing only. */
    reenqueueChangesets?: typeof _reenqueueChangesets
}

export const ReenqueueChangesetsModal: React.FunctionComponent<ReenqueueChangesetsModalProps> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
    reenqueueChangesets = _reenqueueChangesets,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)

    const onSubmit = useCallback<React.FormEventHandler>(async () => {
        setIsLoading(true)
        try {
            const ids = await changesetIDs()
            await reenqueueChangesets(batchChangeID, ids)
            afterCreate()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [changesetIDs, reenqueueChangesets, batchChangeID, afterCreate])

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={LABEL_ID}
        >
            <h3 id={LABEL_ID}>Re-enqueue changesets</h3>
            <p className="mb-4">Are you sure you want to re-enqueue all the selected changesets?</p>
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
                    {isLoading === true && <LoadingSpinner className="icon-inline" />}
                    Re-enqueue
                </button>
            </div>
        </Dialog>
    )
}

const LABEL_ID = 'reenqueue-changesets-modal-title'
