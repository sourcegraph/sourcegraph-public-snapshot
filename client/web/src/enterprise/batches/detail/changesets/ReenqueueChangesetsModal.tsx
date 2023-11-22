import React, { useCallback, useState } from 'react'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, Modal, H3, Text, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import type { Scalars } from '../../../../graphql-operations'
import { reenqueueChangesets as _reenqueueChangesets } from '../backend'

export interface ReenqueueChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]

    /** For testing only. */
    reenqueueChangesets?: typeof _reenqueueChangesets
}

export const ReenqueueChangesetsModal: React.FunctionComponent<
    React.PropsWithChildren<ReenqueueChangesetsModalProps>
> = ({ onCancel, afterCreate, batchChangeID, changesetIDs, reenqueueChangesets = _reenqueueChangesets }) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)

    const onSubmit = useCallback<React.FormEventHandler>(async () => {
        setIsLoading(true)
        try {
            await reenqueueChangesets(batchChangeID, changesetIDs)
            afterCreate()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [changesetIDs, reenqueueChangesets, batchChangeID, afterCreate])

    return (
        <Modal onDismiss={onCancel} aria-labelledby={LABEL_ID}>
            <H3 id={LABEL_ID}>Re-enqueue changesets</H3>
            <Text className="mb-4">Are you sure you want to re-enqueue all the selected changesets?</Text>
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
                    label="Re-enqueue"
                />
            </div>
        </Modal>
    )
}

const LABEL_ID = 'reenqueue-changesets-modal-title'
