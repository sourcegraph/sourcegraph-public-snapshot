import React, { useRef } from 'react'

import { Button, FeedbackPrompt } from '@sourcegraph/wildcard'

import { useHandleSubmitFeedback } from '../../hooks'

interface SourcegraphExtensionFeedbackProps {
    extensionID: string
}

export const SourcegraphExtensionFeedback: React.FunctionComponent<
    React.PropsWithChildren<SourcegraphExtensionFeedbackProps>
> = ({ extensionID }) => {
    const triggerButtonReference = useRef<HTMLButtonElement>(null)
    const textPrefix = `Sourcegraph extension ${extensionID}: `
    const labelId = 'sourcegraph-extension-feedback-modal'

    const { handleSubmitFeedback } = useHandleSubmitFeedback({ textPrefix })

    return (
        <FeedbackPrompt
            triggerButtonReference={triggerButtonReference}
            modal={true}
            modalLabelId={labelId}
            onSubmit={handleSubmitFeedback}
        >
            {({ onClick }) => (
                <Button className="p-0" onClick={onClick} variant="link" ref={triggerButtonReference}>
                    <small>Message the author</small>
                </Button>
            )}
        </FeedbackPrompt>
    )
}
