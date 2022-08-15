import React from 'react'

import classNames from 'classnames'

export const ActionButtons: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    children,
    className,
}) => (
    <div
        className={classNames('d-flex flex-column flex-shrink-0 align-items-center justify-content-center', className)}
    >
        {children}
    </div>
)
