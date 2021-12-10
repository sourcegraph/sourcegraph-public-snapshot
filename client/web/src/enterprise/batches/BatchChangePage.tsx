import React from 'react'

import { SettingsOrgSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings'
import { PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../batches/icons'

const getNamespaceDisplayName = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return namespace.displayName ?? namespace.username
        case 'Org':
            return namespace.displayName ?? namespace.name
    }
}

/** TODO: This duplicates the URL field from the org/user resolvers on the backend, but we
 * don't have access to that from the settings cascade presently. Can we get it included
 * in the cascade instead somehow? */
const getNamespaceBatchChangesURL = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return '/users/' + namespace.username + '/batch-changes'
        case 'Org':
            return '/organizations/' + namespace.name + '/batch-changes'
    }
}

interface BatchChangePageProps {
    /** The namespace that should appear in the topmost `PageHeader`. */
    namespace: SettingsUserSubject | SettingsOrgSubject
    /** The title to use in the topmost `PageHeader`, alongside the `namespaceName`. */
    title: string
    /** The description to use in the topmost `PageHeader` beneath the titles. */
    description?: string | null
    /** Optionally, any action buttons that should appear in the top left of the page. */
    actionButtons?: JSX.Element
}

/**
 * BatchChangePage is a page layout component that renders a consistent header for
 * SSBC-style batch change pages and shoulld wrap the other content contained on the page.
 */
export const BatchChangePage: React.FunctionComponent<BatchChangePageProps> = ({
    children,
    namespace,
    title,
    description,
    actionButtons,
}) => (
    <div className="d-flex flex-column p-4 w-100 h-100">
        <div className="d-flex flex-0 justify-content-between align-items-start">
            <PageHeader
                path={[
                    { icon: BatchChangesIcon },
                    {
                        to: getNamespaceBatchChangesURL(namespace),
                        text: getNamespaceDisplayName(namespace),
                    },
                    { text: title },
                ]}
                className="flex-1 pb-2"
                description={
                    description || 'Run custom code over hundreds of repositories and manage the resulting changesets.'
                }
            />
            <div className="d-flex flex-column flex-0 align-items-center justify-content-center">{actionButtons}</div>
        </div>
        {children}
    </div>
)
