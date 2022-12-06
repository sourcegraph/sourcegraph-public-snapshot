import React, { useCallback, useState } from 'react'

import { mdiCog, mdiDelete } from '@mdi/js'
import { map, mapTo } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Button, H3, Icon, Link, Text, Tooltip } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../backend/graphql'
import { defaultExternalServices } from '../components/externalServices/externalServices'
import { DeleteWebhookResult, DeleteWebhookVariables, ExternalServiceKind, Scalars } from '../graphql-operations'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, isErrorLike } from '@sourcegraph/common'
import styles from './WebhookNode.module.scss'

export interface WebhookProps {
    id: string
    name: string
    codeHostKind: ExternalServiceKind
    codeHostURN: string

    afterDelete: () => void
}

function deleteWebhook(hookID: Scalars['ID']): Promise<void> {
    return requestGraphQL<DeleteWebhookResult, DeleteWebhookVariables>(
        gql`
            mutation DeleteWebhook($hookID: ID!) {
                deleteWebhook(id: $hookID) {
                    alwaysNil
                }
            }
        `,
        { hookID }
    )
        .pipe(map(dataOrThrowErrors), mapTo(undefined))
        .toPromise()
}

export const WebhookNode: React.FunctionComponent<React.PropsWithChildren<WebhookProps>> = ({
    id,
    name,
    codeHostKind,
    codeHostURN,
    afterDelete,
}) => {
    const IconComponent = defaultExternalServices[codeHostKind].icon
    const [isDeleting, setIsDeleting] = useState<boolean | Error>(false)

    const onDelete = useCallback(async () => {
        if (
            !window.confirm(
                'Delete this webhook? Any external webhooks configured to point at this webhook will no longer be received.'
            )
        ) {
            return
        }
        setIsDeleting(true)
        try {
            await deleteWebhook(id)
            setIsDeleting(false)
            if (afterDelete) {
                afterDelete()
            }
        } catch (error) {
            setIsDeleting(asError(error))
        }
    }, [id, afterDelete])

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
                            disabled={isDeleting === true}
                            onClick={onDelete}
                        >
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                        </Button>
                        {isErrorLike(isDeleting) && <ErrorAlert className="mt-2" error={isDeleting} />}
                    </Tooltip>
                </div>
            </div>
        </>
    )
}
