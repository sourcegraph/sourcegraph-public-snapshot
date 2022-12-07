import React from 'react'

import { mdiCog, mdiDelete } from '@mdi/js'
import { noop } from 'lodash'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useMutation } from '@sourcegraph/http-client'
import { Button, H3, Icon, Link, LoadingSpinner, Text, Tooltip } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../components/externalServices/externalServices'
import { DeleteWebhookResult, DeleteWebhookVariables, ExternalServiceKind } from '../graphql-operations'

import { DELETE_WEBHOOK } from './backend'

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

    const [deleteWebhook, { error, loading: isDeleting }] = useMutation<DeleteWebhookResult, DeleteWebhookVariables>(
        DELETE_WEBHOOK,
        { variables: { hookID: id }, onCompleted: () => window.location.reload() }
    )

    return (
        <>
            <span className={styles.nodeSeparator} />
            <div className="pl-1">
                <H3 className="pr-2">
                    {' '}
                    <Link to={`/site-admin/webhooks/${id}`}>{name}</Link>
                    <Text className="mb-0">
                        <small>
                            <Icon inline={true} as={IconComponent} aria-label="Code host logo" className="mr-2" />
                            {codeHostURN}
                        </small>
                    </Text>
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
                        <Button
                            aria-label="Delete"
                            className="test-delete-webhook"
                            variant="danger"
                            size="sm"
                            disabled={isDeleting}
                            onClick={event => {
                                event.preventDefault()
                                deleteWebhook().catch(
                                    // noop here is used because creation error is handled directly when useMutation is called
                                    noop
                                )
                            }}
                        >
                            <>
                                {isDeleting && <LoadingSpinner />}
                                <Icon aria-hidden={true} svgPath={mdiDelete} />
                            </>
                        </Button>
                    </Tooltip>
                </div>
            </div>
            {error && <ErrorAlert className="mt-2" prefix="Error during webhook deletion" error={error} />}
        </>
    )
}
