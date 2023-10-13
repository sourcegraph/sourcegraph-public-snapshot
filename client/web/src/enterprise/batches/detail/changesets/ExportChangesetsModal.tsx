import React, { useCallback } from 'react'

import { Button, Modal, H3, ErrorAlert, Select } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import type { Scalars, GetChangesetsByIDsResult } from '../../../../graphql-operations'
import { useGetChangesetsByIDs } from '../backend'

export interface ExportChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]
}

const exportOptions = {
    CSV: 'csv',
    JSON: 'json',
} as const

type ExportFormat = typeof exportOptions[keyof typeof exportOptions]

const headers = ['title', 'externalURL', 'repository', 'reviewState', 'state'] as const

export const ExportChangesetsModal: React.FunctionComponent<React.PropsWithChildren<ExportChangesetsModalProps>> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
}) => {
    const [getChangesetsByIDs, { loading, error }] = useGetChangesetsByIDs(batchChangeID, changesetIDs)
    const [selectedDataExportType, setSelectedDataExportType] = React.useState<ExportFormat>(exportOptions.CSV)

    const handleFormatChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(event => {
        setSelectedDataExportType(event.target.value as ExportFormat)
    }, [])

    const constructCSVDataExport = useCallback(
        (nodes: GetChangesetsByIDsResult['getChangesetsByIDs']['nodes']): string => {
            const csvRows: (string | null)[] = []
            csvRows.push(headers.join(', '))

            for (const node of nodes) {
                if (node.__typename === 'ExternalChangeset') {
                    // the order is quite important here to ensure the items inserted into `csvRows`
                    // match the headers array above.
                    csvRows.push(
                        [
                            node.title || '',
                            node.externalURL?.url || '',
                            node.repository.name,
                            node.reviewState,
                            node.state,
                        ].join(', ')
                    )
                }
            }
            return csvRows.join('\n')
        },
        []
    )

    const constructJSONDataExport = useCallback(
        (nodes: GetChangesetsByIDsResult['getChangesetsByIDs']['nodes']): string => {
            const jsonRows: Record<typeof headers[number], string | null>[] = []
            for (const node of nodes) {
                if (node.__typename === 'ExternalChangeset') {
                    jsonRows.push({
                        title: node.title,
                        externalURL: node.externalURL?.url || '',
                        repository: node.repository.name,
                        reviewState: node.reviewState,
                        state: node.state,
                    })
                }
            }
            return JSON.stringify(jsonRows, null, 2)
        },
        []
    )

    const onSubmit = useCallback<React.FormEventHandler>(
        () =>
            getChangesetsByIDs({
                onCompleted: node => {
                    let exportData: string
                    if (selectedDataExportType === exportOptions.CSV) {
                        exportData = constructCSVDataExport(node.getChangesetsByIDs.nodes)
                    } else {
                        exportData = constructJSONDataExport(node.getChangesetsByIDs.nodes)
                    }

                    const blob = new Blob([exportData], {
                        type: selectedDataExportType === exportOptions.CSV ? 'text/csv' : 'application/json',
                    })

                    const url = URL.createObjectURL(blob)

                    const element = document.createElement('a')
                    element.download = `batch_change_export.${selectedDataExportType}`
                    element.href = url
                    document.body.append(element)
                    element.click()

                    // cleanup: free memory from the blob URL
                    URL.revokeObjectURL(url)
                    afterCreate()
                },
            }),
        [getChangesetsByIDs, afterCreate, selectedDataExportType, constructCSVDataExport, constructJSONDataExport]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID}>
            <H3 id={MODAL_LABEL_ID}>Export changesets</H3>

            <Select
                id="format"
                label="What format would you like to export the changesets in?"
                message="Only CSV and JSON formats are supported."
                value={selectedDataExportType}
                onChange={handleFormatChange}
            >
                <option value={exportOptions.CSV}>CSV</option>
                <option value={exportOptions.JSON}>JSON</option>
            </Select>

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
