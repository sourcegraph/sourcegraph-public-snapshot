import Dialog from '@reach/dialog'
import React, { useCallback, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../components/alerts'
import { Scalars } from '../../../../graphql-operations'
import { mergeChangesets as _mergeChangesets } from '../backend'

export interface MergeChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: () => Promise<Scalars['ID'][]>

    /** For testing only. */
    mergeChangesets?: typeof _mergeChangesets
}

export const MergeChangesetsModal: React.FunctionComponent<MergeChangesetsModalProps> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
    mergeChangesets = _mergeChangesets,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [squash, setSquash] = useState<boolean>(false)

    const onSubmit = useCallback<React.FormEventHandler>(async () => {
        setIsLoading(true)
        try {
            const ids = await changesetIDs()
            await mergeChangesets(batchChangeID, ids, squash)
            afterCreate()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [changesetIDs, mergeChangesets, batchChangeID, squash, afterCreate])

    const onToggleSquash = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setSquash(event.target.checked)
    }, [])

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={MODAL_LABEL_ID}
        >
            <h3 id={MODAL_LABEL_ID}>Merge changesets</h3>
            <p className="mb-4">Are you sure you want to attempt to merge all the selected changesets?</p>
            <Form>
                <div className="form-group">
                    <div className="form-check">
                        <input
                            id={CHECKBOX_ID}
                            type="checkbox"
                            checked={squash}
                            onChange={onToggleSquash}
                            className="form-check-input"
                            disabled={isLoading === true}
                        />
                        <label className="form-check-label" htmlFor={CHECKBOX_ID}>
                            Squash merge all selected changesets.
                        </label>
                    </div>
                </div>
            </Form>
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
                    Merge
                </button>
            </div>
        </Dialog>
    )
}

const MODAL_LABEL_ID = 'merge-changesets-modal-title'
const CHECKBOX_ID = 'merge-changesets-modal-squash-check'
