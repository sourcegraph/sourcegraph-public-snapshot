import classNames from 'classnames'
import React from 'react'

interface Props {
    title: string
    listTag: 'ol' | 'ul'
    className?: string
}

export const ScrollListRow: React.FunctionComponent<Props> = ({
    title,
    listTag: ListTag = 'ul',
    className,
    children,
}) => (
    <div className={classNames('', className)}>
        <h4>{title}</h4>
        <div
            className="overflow-auto rounded border-left border-right"
            style={{ width: 'fit-content', maxWidth: '100%' }}
        >
            <ListTag className={classNames('list-group list-group-horizontal')}>{children}</ListTag>
        </div>
    </div>
)
