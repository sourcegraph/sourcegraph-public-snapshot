import React, { useCallback, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, Modal } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import { Scalars } from '../../../../graphql-operations'
import { mergeChangesets as _mergeChangesets } from '../backend'

export interface MergeChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]

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
            await mergeChangesets(batchChangeID, changesetIDs, squash)
            afterCreate()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [changesetIDs, mergeChangesets, batchChangeID, squash, afterCreate])

    const onToggleSquash = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setSquash(event.target.checked)
    }, [])

    return (
        <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID}>
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
                <Button
                    disabled={isLoading === true}
                    className="mr-2"
                    onClick={onCancel}
                    outline={true}
                    variant="secondary"
                >
                    Cancel
                </Button>
                <LoaderButton
                    onClick={onSubmit}
                    disabled={isLoading === true}
                    variant="primary"
                    loading={isLoading === true}
                    alwaysShowLabel={true}
                    label="Merge"
                />
            </div>
        </Modal>
    )
}

const MODAL_LABEL_ID = 'merge-changesets-modal-title'
const CHECKBOX_ID = 'merge-changesets-modal-squash-check'
