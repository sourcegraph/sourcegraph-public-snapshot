import Dialog from '@reach/dialog'
import React, { useState } from 'react'

import { FeedbackPromptContent } from '../../nav/Feedback/FeedbackPrompt'

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
            <button type="button" className="btn btn-sm btn-link p-0" onClick={toggleIsOpen}>
                Message the author
            </button>
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
