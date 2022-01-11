import Dialog from '@reach/dialog'
import React, { useState } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { FeedbackPromptContent } from '../../nav/Feedback'

interface SourcegraphExtensionFeedbackProps {
    extensionID: string
}

export const SourcegraphExtensionFeedback: React.FunctionComponent<SourcegraphExtensionFeedbackProps> = ({
    extensionID,
}) => {
    const [isOpen, setIsOpen] = useState(false)

    const toggleIsOpen = (): void => setIsOpen(!isOpen)
    const onClose = (): void => setIsOpen(false)
    const textPrefix = `Sourcegraph extension ${extensionID}: `
    const labelId = 'sourcegraph-extension-feedback-modal'

    return (
        <>
            <Button className="p-0" onClick={toggleIsOpen} variant="link">
                <small>Message the author</small>
            </Button>
            {isOpen && (
                <Dialog
                    className="modal-body modal-body--top-third p-4 rounded border"
                    onDismiss={onClose}
                    aria-labelledby={labelId}
                >
                    <FeedbackPromptContent closePrompt={onClose} textPrefix={textPrefix} />
                </Dialog>
            )}
        </>
    )
}
