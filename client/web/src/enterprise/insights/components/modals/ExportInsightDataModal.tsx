import type { FC } from 'react'

import { escapeRegExp } from 'lodash'

import { Modal, Text, H2, Link } from '@sourcegraph/wildcard'

import { DownloadFileButton } from '../../../../components/DownloadFileButton'

interface ExportInsightDataModalProps {
    insightId: string
    insightTitle: string
    showModal: boolean
    onCancel: () => void
    onConfirm: () => void
}

export const ExportInsightDataModal: FC<ExportInsightDataModalProps> = props => {
    const { insightId, insightTitle, showModal, onCancel, onConfirm } = props

    return (
        <Modal isOpen={showModal} position="center" aria-label="Export insight data modal" onDismiss={onCancel}>
            <H2 className="font-weight-normal">Export data for '{insightTitle}' insight?</H2>

            <Text className="mt-4 mb-2">
                This will create a CSV archive of all data for this Code Insight, including
                <Link to="/help/code_insights/explanations/data_retention" target="_blank" rel="noopener">
                    {' '}
                    data that has been archived
                </Link>
                .
            </Text>
            <Text>This will only include data that you are permitted to see.</Text>
            <div className="d-flex justify-content-end mt-5">
                <DownloadFileButton
                    fileName={escapeRegExp(insightTitle)}
                    fileUrl={`/.api/insights/export/${insightId}`}
                    variant="primary"
                    onClick={onConfirm}
                >
                    Export data as CSV
                </DownloadFileButton>
            </div>
        </Modal>
    )
}
