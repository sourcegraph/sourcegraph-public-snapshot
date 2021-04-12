import classNames from 'classnames'
import * as React from 'react'

export const ShowMoreButton: React.FunctionComponent<{ onClick: () => void; className?: string }> = ({
    onClick,
    className,
}) => (
    <div className="text-center py-3">
        <button type="button" className={classNames('btn btn-link', className)} onClick={onClick}>
            Show more
        </button>
    </div>
)
