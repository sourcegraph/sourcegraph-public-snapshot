import React from 'react'

import { Alert } from '@sourcegraph/wildcard'

interface FlashMessageProps {
    state: string
    message: string
    className?: string
}

export const FlashMessage: React.FunctionComponent<FlashMessageProps> = ({ state, message, className }) => (
    <Alert variant={state === 'SUCCESS' ? 'success' : 'warning'} className={className}>
        {message}
    </Alert>
)
