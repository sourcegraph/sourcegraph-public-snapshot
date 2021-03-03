import React from 'react'
import { StatusBadge } from '../../components/StatusBadge'

/**
 * Shows a "WIP" badge for extensions.
 */
export const ExtensionStatusBadge: React.FunctionComponent<{ viewerCanAdminister: boolean }> = ({
    viewerCanAdminister,
}) => (
    <StatusBadge
        status="wip"
        tooltip={
            viewerCanAdminister
                ? 'Remove "WIP" from the title when this extension is ready for use.'
                : 'Work in progress (not ready for use)'
        }
    />
)
