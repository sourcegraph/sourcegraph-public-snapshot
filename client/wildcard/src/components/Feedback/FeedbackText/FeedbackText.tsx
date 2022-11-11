import * as React from 'react'

import { Link } from '../../Link'
import { Text } from '../../Typography'

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
export const FeedbackText: React.FunctionComponent<React.PropsWithChildren<FeedbackTextProps>> = ({
    className,
    footerText,
    headerText,
}) => (
    <Text className={className}>
        {headerText || 'Questions/feedback?'} Contact us at{' '}
        <Link to="https://twitter.com/sourcegraph" target="_blank" rel="noopener noreferrer">
            @sourcegraph
        </Link>{' '}
        or{' '}
        <Link to="mailto:support@sourcegraph.com" target="_blank" rel="noopener noreferrer">
            support@sourcegraph.com
        </Link>
        , or file issues on our{' '}
        <Link to="https://github.com/sourcegraph/issues/issues" target="_blank" rel="noopener noreferrer">
            public issue tracker
        </Link>
        . {footerText}
    </Text>
)
