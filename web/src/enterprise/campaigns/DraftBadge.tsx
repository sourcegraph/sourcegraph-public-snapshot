import React from 'react'
import classNames from 'classnames'

interface Props {
    className?: string
}

export const DraftBadge: React.FunctionComponent<Props> = ({ className }) => (
    <span className={classNames('badge badge-info', className)}>Draft</span>
)
