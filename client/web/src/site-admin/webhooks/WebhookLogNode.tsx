import React from 'react'

import { WebhookLogFields } from '../../graphql-operations'

export interface Props {
    node: WebhookLogFields
}

export const WebhookLogNode: React.FunctionComponent<Props> = ({ node }) => {
    Math.min(1, 2)

    return (
        <>
            <span>&lt;</span>
            <span>{node.statusCode}</span>
            <span>{node.externalService ? node.externalService.displayName : 'Unmatched'}</span>
            <span>{node.receivedAt}</span>
        </>
    )
}
