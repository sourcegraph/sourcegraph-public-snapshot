import React from 'react'

import { mdiCog, mdiDelete } from '@mdi/js'

import { Button, H3, Icon, Tooltip } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../components/externalServices/externalServices'
import { ExternalServiceKind } from '../graphql-operations'

import styles from './WebhookNode.module.scss'

export interface WebhookProps {
    codeHostKind: ExternalServiceKind
    codeHostURN: string
}

export const WebhookNode: React.FunctionComponent<React.PropsWithChildren<WebhookProps>> = ({
    codeHostKind,
    codeHostURN,
}) => {
    const IconComponent = defaultExternalServices[codeHostKind].icon
    return (
        <>
            <span className={styles.nodeSeparator} />
            <div className="pl-1">
                <H3 className="pr-2">
                    {' '}
                    <Icon inline={true} as={IconComponent} aria-label="Code host logo" className="mr-2" />
                    {codeHostURN}
                </H3>
            </div>
            <div className="d-flex flex-shrink-0 ml-3">
                <div>
                    <Tooltip content="Edit webhook">
                        <Button aria-label="Edit" className="test-edit-webhook" variant="secondary" size="sm">
                            <Icon aria-hidden={true} svgPath={mdiCog} /> Edit
                        </Button>
                    </Tooltip>
                </div>
                <div className="ml-1">
                    <Tooltip content="Delete webhook">
                        <Button aria-label="Delete" className="test-delete-webhook" variant="danger" size="sm">
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                        </Button>
                    </Tooltip>
                </div>
            </div>
        </>
    )
}
