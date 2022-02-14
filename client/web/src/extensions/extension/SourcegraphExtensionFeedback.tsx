import React from 'react'

import { Button, FeedbackPrompt } from '@sourcegraph/wildcard'

import { useHandleSubmitFeedback } from '../../hooks'

interface SourcegraphExtensionFeedbackProps {
    extensionID: string
}

export const SourcegraphExtensionFeedback: React.FunctionComponent<SourcegraphExtensionFeedbackProps> = ({
    extensionID,
}) => {
    const textPrefix = `Sourcegraph extension ${extensionID}: `
    const { handleSubmitFeedback } = useHandleSubmitFeedback(undefined, textPrefix)
    const labelId = 'sourcegraph-extension-feedback-modal'

    return (
        <FeedbackPrompt modal={true} modalLabelId={labelId} onSubmit={handleSubmitFeedback}>
            {({ onClick }) => (
                <Button className="p-0" onClick={onClick} variant="link">
                    <small>Message the author</small>
                </Button>
            )}
        </FeedbackPrompt>
    )
}
