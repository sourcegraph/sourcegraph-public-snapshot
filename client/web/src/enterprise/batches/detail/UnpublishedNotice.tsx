import React from 'react'

import { pluralize } from '@sourcegraph/common'
import { AlertLink, Alert } from '@sourcegraph/wildcard'

interface UnpublishedNoticeProps {
    unpublished: number
    total: number
    className?: string
}

export const UnpublishedNotice: React.FunctionComponent<React.PropsWithChildren<UnpublishedNoticeProps>> = ({
    unpublished,
    total,
    className,
}) => {
    if (total === 0 || unpublished !== total) {
        return <></>
    }
    return (
        <Alert className={className} variant="secondary">
            {unpublished} unpublished {pluralize('changeset', unpublished, 'changesets')}. Select changeset(s) and
            choose the 'Publish changesets' action to publish them, or{' '}
            <AlertLink
                to="/help/batch_changes/how-tos/publishing_changesets#publishing-changesets"
                rel="noopener"
                target="_blank"
            >
                read more about publishing changesets
            </AlertLink>
            .
        </Alert>
    )
}
