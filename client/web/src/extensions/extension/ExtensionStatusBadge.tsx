import React from 'react'

import classNames from 'classnames'

import { ProductStatusBadge } from '@sourcegraph/wildcard'

/**
 * Shows an "EXPERIMENTAL" badge for work-in-progress extensions.
 */
export const ExtensionStatusBadge: React.FunctionComponent<
    React.PropsWithChildren<{ viewerCanAdminister: boolean; className?: string }>
> = ({ viewerCanAdminister, className }) => (
    <ProductStatusBadge
        status="experimental"
        tooltip={
            viewerCanAdminister
                ? 'Remove "WIP" from the manifest when this extension is ready for use.'
                : 'Work in progress (not ready for use)'
        }
        className={classNames('text-uppercase', className)}
    />
)
