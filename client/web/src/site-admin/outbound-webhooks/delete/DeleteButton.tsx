import type { FC } from 'react'

import { mdiAlertCircle, mdiDelete } from '@mdi/js'
import { noop } from 'lodash'

import { useMutation } from '@sourcegraph/http-client'
import { Button, Icon, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import type { DeleteOutboundWebhookResult, DeleteOutboundWebhookVariables } from '../../../graphql-operations'
import { DELETE_OUTBOUND_WEBHOOK } from '../backend'

export interface DeleteButtonProps {
    className?: string
    id: string
    onDeleted: () => void
}

export const DeleteButton: FC<DeleteButtonProps> = ({ className, id, onDeleted }) => {
    const [deleteOutboundWebhook, { error, loading }] = useMutation<
        DeleteOutboundWebhookResult,
        DeleteOutboundWebhookVariables
    >(DELETE_OUTBOUND_WEBHOOK, { variables: { id }, onCompleted: () => onDeleted() })

    if (error) {
        return (
            <Tooltip content={error.message}>
                <Button variant="danger" className={className} disabled={true}>
                    <Icon aria-label={error.message} svgPath={mdiAlertCircle} />
                </Button>
            </Tooltip>
        )
    }

    if (loading) {
        return (
            <Button variant="danger" className={className} disabled={true}>
                <LoadingSpinner />
            </Button>
        )
    }

    return (
        <Button
            variant="danger"
            className={className}
            onClick={event => {
                event.preventDefault()
                if (!window.confirm('Are you sure you want to delete this outbound webhook?')) {
                    return
                }
                deleteOutboundWebhook().catch(noop)
            }}
        >
            <Icon aria-hidden={true} svgPath={mdiDelete} /> Delete
        </Button>
    )
}
