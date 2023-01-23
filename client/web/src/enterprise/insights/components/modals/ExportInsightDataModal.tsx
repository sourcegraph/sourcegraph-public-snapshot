import { FC } from 'react'

import { Button, Modal, Text, H2 } from '@sourcegraph/wildcard'

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
                This will create a CVS archive of all data for this Code Insight, including data that has been archived.
            </Text>
            <Text>This will only include data that you are permitted to see.</Text>
            <div className="d-flex justify-content-end mt-5">
                <Button
                    as="a"
                    href={`/.api/insights/export/${insightId}`}
                    autoFocus={true}
                    download={true}
                    variant="primary"
                    onClick={onConfirm}
                >
                    Export data as CSV
                </Button>
            </div>
        </Modal>
    )
}
