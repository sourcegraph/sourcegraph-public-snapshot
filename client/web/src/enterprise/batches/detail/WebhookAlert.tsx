import classNames from 'classnames'
import React, { useCallback, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'

import { BatchChangeFields } from '../../../graphql-operations'

import styles from './WebhookAlert.module.scss'

export interface Props {
    batchChange: Pick<
        BatchChangeFields,
        'id' | 'externalServicesWithoutWebhooks' | 'hasExternalServicesWithoutWebhooks'
    >
}

export const WebhookAlert: React.FunctionComponent<Props> = ({
    batchChange: { id, externalServicesWithoutWebhooks, hasExternalServicesWithoutWebhooks },
}) => {
    const [open, setOpen] = useState(false)
    const toggleOpen = useCallback(() => setOpen(!open), [open])

    if (window.context.batchChangesDisableWebhooksWarning) {
        return null
    }

    if (!hasExternalServicesWithoutWebhooks) {
        return null
    }

    if (!externalServicesWithoutWebhooks) {
        return (
            <DismissibleAlert className="alert-warning" partialStorageKey={id}>
                <div>
                    <h4>Changeset information may not be up to date</h4>
                    Sourcegraph will poll for updates because one or more code hosts are not configured to use webhooks.
                    Ask your site admin to configure webhooks.
                </div>
            </DismissibleAlert>
        )
    }

    return (
        <DismissibleAlert className="alert-warning" partialStorageKey={id}>
            <div>
                <h4>Changeset information may not be up to date</h4>
                <p className={styles.blurb}>
                    Sourcegraph will poll for updates because{' '}
                    <button
                        type="button"
                        className={classNames(styles.openLink, 'btn btn-link p-0')}
                        onClick={toggleOpen}
                    >
                        one or more code hosts
                    </button>{' '}
                    are not configured to use webhooks. Learn how to{' '}
                    <Link to="https://docs.sourcegraph.com/batch_changes/how-tos/site_admin_configuration">
                        configure webooks
                    </Link>{' '}
                    or disable this warning.{' '}
                </p>
                {open && (
                    <ul>
                        {externalServicesWithoutWebhooks.nodes.map(({ id, displayName }) => (
                            <li key={id}>
                                <Link to={`/site-admin/external-services/${id}`}>{displayName}</Link>
                            </li>
                        ))}
                        {externalServicesWithoutWebhooks.pageInfo.hasNextPage && <li key="...">(and more)</li>}
                    </ul>
                )}
            </div>
        </DismissibleAlert>
    )
}
