import React from 'react'

import { H3, Icon, Link, Text } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../components/externalServices/externalServices'
import type { ExternalServiceKind } from '../graphql-operations'

import styles from './WebhookNode.module.scss'

export interface WebhookProps {
    id: string
    name: string
    codeHostKind: ExternalServiceKind
    codeHostURN: string
}

export const WebhookNode: React.FunctionComponent<React.PropsWithChildren<WebhookProps>> = ({
    id,
    name,
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
                    <Link to={`/site-admin/webhooks/incoming/${id}`}>{name}</Link>
                    <Text className="mb-0">
                        <small>
                            <Icon inline={true} as={IconComponent} aria-label="Code host logo" className="mr-2" />
                            {codeHostURN}
                        </small>
                    </Text>
                </H3>
            </div>
        </>
    )
}
