import classNames from 'classnames'
import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

export const BatchChangeStateBadge: React.FunctionComponent<{ isClosed: boolean; className?: string }> = ({
    isClosed,
    className,
}) => {
    if (isClosed) {
        return (
            <Badge variant="danger" className={classNames('text-uppercase', className)}>
                Closed
            </Badge>
        )
    }
    return (
        <Badge variant="success" className={classNames('text-uppercase', className)}>
            Open
        </Badge>
    )
}
