import React, { useCallback, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Modal, Typography } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import { Scalars } from '../../../../graphql-operations'
import { detachChangesets as _detachChangesets } from '../backend'

export interface DetachChangesetsModalProps extends TelemetryProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]

    /** For testing only. */
    detachChangesets?: typeof _detachChangesets
}

export const DetachChangesetsModal: React.FunctionComponent<React.PropsWithChildren<DetachChangesetsModalProps>> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
    telemetryService,
    detachChangesets = _detachChangesets,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)

    const onSubmit = useCallback<React.FormEventHandler>(async () => {
        setIsLoading(true)
        try {
            await detachChangesets(batchChangeID, changesetIDs)
            telemetryService.logViewEvent('BatchChangeDetailsPageDetachArchivedChangesets')
            afterCreate()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [changesetIDs, detachChangesets, batchChangeID, telemetryService, afterCreate])

    const labelId = 'detach-changesets-modal-title'

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <Typography.H3 id={labelId}>Detach changesets</Typography.H3>
            <p className="mb-4">Are you sure you want to detach the selected changesets?</p>
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
                    label="Detach"
                />
            </div>
        </Modal>
    )
}
