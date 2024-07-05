import React, { useCallback } from 'react'

import { logger } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, H3, Modal, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../components/LoaderButton'
import { WebhookFields } from '../graphql-operations'

import { useDeleteWebhook } from './backend'

export interface WebhookConfirmDeleteModalProps extends TelemetryV2Props {
    webhook: WebhookFields

    onCancel: () => void
    afterDelete: () => void
}

export const WebhookConfirmDeleteModal: React.FunctionComponent<
    React.PropsWithChildren<WebhookConfirmDeleteModalProps>
> = ({ webhook, onCancel, afterDelete, telemetryRecorder }) => {
    const labelId = 'deleteWebhook'

    const [deleteWebhook, { loading, error }] = useDeleteWebhook()

    const onDelete = useCallback<React.MouseEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await deleteWebhook({ variables: { id: webhook.id } })

                telemetryRecorder.recordEvent('webhook', 'delete')
                afterDelete()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
                telemetryRecorder.recordEvent('webhook', 'deleteFail')
            }
        },
        [deleteWebhook, webhook.id, telemetryRecorder, afterDelete]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Delete webhook {webhook.name}?</H3>

            <strong className="d-block text-danger my-3">
                Removing webhooks is irreversible and all incoming webhooks will be rejected.
            </strong>

            {error && <ErrorAlert error={error} />}

            <div className="d-flex justify-content-end pt-1">
                <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                    Cancel
                </Button>
                <LoaderButton
                    disabled={loading}
                    onClick={onDelete}
                    variant="danger"
                    loading={loading}
                    alwaysShowLabel={true}
                    label="Delete webhook"
                />
            </div>
        </Modal>
    )
}
