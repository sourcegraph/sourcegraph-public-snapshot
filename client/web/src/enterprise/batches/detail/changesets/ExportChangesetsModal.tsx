import React, { useCallback, useState } from 'react'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, Checkbox, Modal, H3, Text, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import { Scalars } from '../../../../graphql-operations'
import { exportChangesets as _exportChangesets } from '../backend'

export interface ExportChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]

    /** For testing only. */
    exportChangesets?: typeof _exportChangesets
}

export const ExportChangesetsModal: React.FunctionComponent<React.PropsWithChildren<ExportChangesetsModalProps>> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
    exportChangesets: exportChangesets = _exportChangesets,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [draft, setDraft] = useState(false)

    const onSubmit = useCallback<React.FormEventHandler>(async () => {
        setIsLoading(true)
        try {
            await exportChangesets(batchChangeID, changesetIDs, draft)
            afterCreate()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [changesetIDs, exportChangesets, batchChangeID, draft, afterCreate])

    return (
        <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID}>
            <H3 id={MODAL_LABEL_ID}>Export changesets</H3>
            <Text className="mb-4">Are you sure you want to export the selected changesets?</Text>
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
                    label="Export"
                />
            </div>
        </Modal>
    )
}

const MODAL_LABEL_ID = 'export-changesets-modal-title'
