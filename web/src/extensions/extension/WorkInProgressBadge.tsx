import React from 'react'

/**
 * Shows a "Work in progress" badge for extensions.
 */
export const WorkInProgressBadge: React.FunctionComponent<{ viewerCanAdminister: boolean }> = ({
    viewerCanAdminister,
}) => (
    <span
        className="badge badge-danger mr-2"
        data-tooltip={
            viewerCanAdminister
                ? 'Remove "WIP" from the title when this extension is ready for use.'
                : 'This extension is still under development and is not ready for use.'
        }
    >
        Work in progress
    </span>
)
