import React from 'react'

import { WebhookLogMessageFields } from '../../graphql-operations'

export interface Props {
    className?: string
    message: WebhookLogMessageFields
}

export const MessagePanel: React.FunctionComponent<Props> = ({ className, message }) => (
    <div className={className}>headers; body</div>
)
