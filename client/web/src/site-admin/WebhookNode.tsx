import React, { useCallback } from 'react'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, ButtonLink, H3, Icon, Text } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../components/externalServices/externalServices'
import type { WebhookFields } from '../graphql-operations'

import { WebhookConfirmDeleteModal } from './WebhookConfirmDeleteModal'

import styles from './WebhookNode.module.scss'

export interface WebhookProps extends TelemetryV2Props {
    webhook: WebhookFields
    first: boolean
    afterDelete: () => void
}

export const WebhookNode: React.FunctionComponent<React.PropsWithChildren<WebhookProps>> = ({
    webhook,
    first,
    afterDelete,
    telemetryRecorder,
}) => {
    const IconComponent = defaultExternalServices[webhook.codeHostKind].icon
    const [showDeleteModal, setShowDeleteModal] = React.useState(false)
    const deleteWebhook = useCallback(() => {
        setShowDeleteModal(true)
    }, [])

    return (
        <>
            {showDeleteModal && (
                <WebhookConfirmDeleteModal
                    webhook={webhook}
                    onCancel={() => setShowDeleteModal(false)}
                    afterDelete={afterDelete}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
            {!first && <span className={styles.nodeSeparator} />}
            <div className="d-flex align-items-center justify-content-between">
                <div className="pl-1">
                    <H3>{webhook.name}</H3>
                    <Text className="mb-0">
                        <Icon inline={true} as={IconComponent} aria-label="Code host logo" className="mr-1" />
                        {webhook.codeHostURN}
                    </Text>
                </div>
                <div>
                    <ButtonLink
                        variant="secondary"
                        to={`/site-admin/webhooks/incoming/${webhook.id}`}
                        className="mr-2"
                        disabled={showDeleteModal}
                    >
                        Edit
                    </ButtonLink>
                    <Button variant="danger" onClick={deleteWebhook} disabled={showDeleteModal}>
                        Delete
                    </Button>
                </div>
            </div>
        </>
    )
}
