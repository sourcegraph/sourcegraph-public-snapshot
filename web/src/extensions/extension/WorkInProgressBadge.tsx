import React from 'react'

/**
 * Shows a "WIP" badge for extensions.
 */
export const WorkInProgressBadge: React.FunctionComponent<{ viewerCanAdminister: boolean }> = ({
    viewerCanAdminister,
}) => (
    <span
        className="badge badge-warning mr-2 text-white"
        data-tooltip={
            viewerCanAdminister
                ? 'Remove "WIP" from the title when this extension is ready for use.'
                : 'Work in progress (not ready for use)'
        }
    >
        WIP
    </span>
)
