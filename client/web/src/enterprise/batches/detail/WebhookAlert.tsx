import React, { useCallback, useState } from 'react'

import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { Button, useObservable, Link, H4, Text } from '@sourcegraph/wildcard'

import { authenticatedUser } from '../../../auth'
import { DismissibleAlert } from '../../../components/DismissibleAlert'
import type { BatchChangeFields } from '../../../graphql-operations'
import { CodeHost } from '../CodeHost'

import styles from './WebhookAlert.module.scss'

export interface Props {
    batchChange: Pick<BatchChangeFields, 'id' | 'currentSpec'>

    // isSiteAdmin is only here for storybook purposes.
    isSiteAdmin?: boolean
}

export const WebhookAlert: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    batchChange: {
        id,
        currentSpec: {
            codeHostsWithoutWebhooks: {
                nodes,
                pageInfo: { hasNextPage },
                totalCount,
            },
        },
    },
    isSiteAdmin,
}) => {
    const user = useObservable(authenticatedUser)
    if (isSiteAdmin === undefined) {
        isSiteAdmin = user?.siteAdmin === true
    }

    const [open, setOpen] = useState(false)
    const toggleOpen = useCallback(() => setOpen(!open), [open])

    if (window.context.batchChangesDisableWebhooksWarning) {
        return null
    }

    if (totalCount === 0) {
        return null
    }

    const SITE_ADMIN_CONFIG_DOC_URL = '/help/batch_changes/how-tos/site_admin_configuration'

    return (
        <DismissibleAlert variant="warning" partialStorageKey={id}>
            <div>
                <H4>Changeset information may not be up to date</H4>
                <Text className={styles.blurb}>
                    Sourcegraph will poll for updates because{' '}
                    <Button className={classNames(styles.openLink, 'p-0')} onClick={toggleOpen} variant="link">
                        {totalCount}{' '}
                        {pluralize('code host is not configured', totalCount, 'code hosts are not configured')}
                    </Button>{' '}
                    to use webhooks.{' '}
                    {isSiteAdmin ? (
                        <>
                            Learn how to <Link to={SITE_ADMIN_CONFIG_DOC_URL}>configure webhooks</Link> or disable this
                            warning.
                        </>
                    ) : (
                        <>
                            Ask your site admin <Link to={SITE_ADMIN_CONFIG_DOC_URL}>to configure webhooks</Link>.
                        </>
                    )}
                </Text>
                {open && (
                    <ul>
                        {nodes.map(codeHost => (
                            <li key={codeHost.externalServiceKind + codeHost.externalServiceURL}>
                                <CodeHost {...codeHost} />
                            </li>
                        ))}
                        {hasNextPage && <li key="and-more">and {totalCount - nodes.length} more</li>}
                    </ul>
                )}
            </div>
        </DismissibleAlert>
    )
}
