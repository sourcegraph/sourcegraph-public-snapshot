import React from 'react'

import { mdiCog, mdiDelete } from '@mdi/js'

import { Button, H3, Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../components/externalServices/externalServices'
import { ExternalServiceKind } from '../graphql-operations'

import styles from './WebhookNode.module.scss'

export interface WebhookProps {
    id: string
    codeHostKind: ExternalServiceKind
    codeHostURN: string
}

export const WebhookNode: React.FunctionComponent<React.PropsWithChildren<WebhookProps>> = ({
    id,
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
                    <Link to={`/site-admin/webhooks/${id}`}>{codeHostURN}</Link>
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
