import React, { useCallback, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, Modal, Typography } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
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

export const CloseChangesetsModal: React.FunctionComponent<React.PropsWithChildren<CloseChangesetsModalProps>> = ({
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
        <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID}>
            <Typography.H3 id={MODAL_LABEL_ID}>Close changesets</Typography.H3>
            <p className="mb-4">Are you sure you want to close all the selected changesets on the code hosts?</p>
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
                    label="Close"
                />
            </div>
        </Modal>
    )
}

const MODAL_LABEL_ID = 'close-changesets-modal-title'
