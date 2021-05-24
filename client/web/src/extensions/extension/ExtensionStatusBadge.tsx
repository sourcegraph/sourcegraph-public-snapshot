import classnames from 'classnames'
import React from 'react'

import { Badge } from '../../components/Badge'

/**
 * Shows an "EXPERIMENTAL" badge for work-in-progress extensions.
 */
export const ExtensionStatusBadge: React.FunctionComponent<{ viewerCanAdminister: boolean; className?: string }> = ({
    viewerCanAdminister,
    className,
}) => (
    <Badge
        status="experimental"
        tooltip={
            viewerCanAdminister
                ? 'Remove "WIP" from the manifest when this extension is ready for use.'
                : 'Work in progress (not ready for use)'
        }
        className={classnames('text-uppercase', className)}
    />
)
