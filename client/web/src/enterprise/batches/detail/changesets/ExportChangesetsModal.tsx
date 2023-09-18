import React, { useCallback } from 'react'

import { Button, Modal, H3, Text, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import { type Scalars } from '../../../../graphql-operations'
import { useExportChangesets } from '../backend'

export interface ExportChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]
}

export const ExportChangesetsModal: React.FunctionComponent<React.PropsWithChildren<ExportChangesetsModalProps>> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
}) => {
    const [exportChangesets, { loading, error }] = useExportChangesets(batchChangeID, changesetIDs)

    const onSubmit = useCallback<React.FormEventHandler>(async () => {
        await exportChangesets({
            onCompleted: data => {
                const { data: csvData, batchChange } = data.exportChangesets
                const blob = new Blob([csvData], { type: 'text/csv' })

                const url = URL.createObjectURL(blob)

                const element = document.createElement('a')
                element.download = `${batchChange}.csv`
                element.href = url
                document.body.append(element)
                element.click()

                // cleanup: free memory from the blob URL
                URL.revokeObjectURL(url)

                afterCreate()
            },
        })
    }, [exportChangesets, afterCreate])

    return (
        <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID}>
            <H3 id={MODAL_LABEL_ID}>Export changesets</H3>
            <Text className="mb-4">Are you sure you want to export the selected changesets?</Text>
            {error && <ErrorAlert error={error} />}
            <div className="d-flex justify-content-end">
                <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                    Cancel
                </Button>
                <LoaderButton
                    onClick={onSubmit}
                    disabled={loading}
                    variant="primary"
                    loading={loading}
                    alwaysShowLabel={true}
                    label="Export"
                />
            </div>
        </Modal>
    )
}

const MODAL_LABEL_ID = 'export-changesets-modal'
