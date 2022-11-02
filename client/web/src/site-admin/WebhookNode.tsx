import React from 'react'

import { mdiCog, mdiDelete } from '@mdi/js'

import { Button, H3, Icon, Text, Tooltip } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../components/externalServices/externalServices'
import { WebhookFields } from '../graphql-operations'

import styles from './WebhookNode.module.scss'

export interface WebhookProps {
    node: WebhookFields
}

export const WebhookNode: React.FunctionComponent<React.PropsWithChildren<WebhookProps>> = ({ node }) => {
    const IconComponent = defaultExternalServices[node.codeHostKind].icon
    return (
        <>
            <span className={styles.nodeSeparator} />
            <div className="pl-1">
                <H3 className="pr-2">{node.codeHostURN}</H3>
                <Text className="mb-0 text-muted">
                    <small>
                        <Icon as={IconComponent} aria-label="Code host logo" className="mr-2" />
                    </small>
                </Text>
            </div>
            <div className="d-flex flex-shrink-0 ml-3">
                <div>
                    <Tooltip content="Edit webhook">
                        <Button
                            aria-label="Edit"
                            className="test-edit-webhook"
                            variant="secondary"
                            size="sm"
                        >
                            <Icon aria-hidden={true} svgPath={mdiCog} /> Edit
                        </Button>
                    </Tooltip>
                </div>
                <div className="ml-1">
                    <Tooltip content="Delete code host connection">
                        <Button
                            aria-label="Delete"
                            className="test-delete-webhook"
                            variant="danger"
                            size="sm"
                        >
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                        </Button>
                    </Tooltip>
                </div>
            </div>
        </>
    )
}
