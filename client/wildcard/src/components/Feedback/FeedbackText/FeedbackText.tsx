import * as React from 'react'

interface FeedbackTextProps {
    /**
     * @default "Questions/feedback?"
     */
    headerText?: React.ReactNode
    footerText?: React.ReactNode
    className?: string
}

/**
 * An abstract UI component which renders a text for feedback.
 */
export const FeedbackText: React.FunctionComponent<FeedbackTextProps> = ({ className, footerText, headerText }) => (
    <p className={className}>
        {headerText || 'Questions/feedback?'} Contact us at{' '}
        <a href="https://twitter.com/sourcegraph" target="_blank" rel="noopener noreferrer">
            @sourcegraph
        </a>{' '}
        or{' '}
        <a href="mailto:support@sourcegraph.com" target="_blank" rel="noopener noreferrer">
            support@sourcegraph.com
        </a>
        , or file issues on our{' '}
        <a href="https://github.com/sourcegraph/issues/issues" target="_blank" rel="noopener noreferrer">
            public issue tracker
        </a>
        . {footerText}
    </p>
)
