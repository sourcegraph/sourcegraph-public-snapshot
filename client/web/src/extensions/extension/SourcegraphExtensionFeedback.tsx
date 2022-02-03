import React, { useState } from 'react'

import { Button, FeedbackPrompt, PopoverTrigger } from '@sourcegraph/wildcard'

import { useHandleSubmitFeedback } from '../../hooks'

interface SourcegraphExtensionFeedbackProps {
    extensionID: string
}

export const SourcegraphExtensionFeedback: React.FunctionComponent<SourcegraphExtensionFeedbackProps> = ({
    extensionID,
}) => {
    const textPrefix = `Sourcegraph extension ${extensionID}: `
    const feedbackSubmitState = useHandleSubmitFeedback({ textPrefix })
    const [isOpen, setIsOpen] = useState(false)

    const toggleIsOpen = (): void => setIsOpen(!isOpen)
    const onClose = (): void => setIsOpen(false)

    return (
        <>
            <FeedbackPrompt closePrompt={onClose} {...feedbackSubmitState} open={isOpen}>
                <PopoverTrigger className="p-0" onClick={toggleIsOpen} as={Button} variant="link">
                    <small>Message the author</small>
                </PopoverTrigger>
            </FeedbackPrompt>
        </>
    )
}
