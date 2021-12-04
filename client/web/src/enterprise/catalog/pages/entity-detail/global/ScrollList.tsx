import classNames from 'classnames'
import React from 'react'

interface Props {
    title: string
    listTag: 'ol' | 'ul'
    className?: string
}

export const ScrollListHorizontal: React.FunctionComponent<Props> = ({
    title,
    listTag: ListTag = 'ul',
    className,
    children,
}) => (
    <div className={classNames(className)}>
        <h4 className="font-weight-bold">{title}</h4>
        <div
            className="overflow-auto rounded border-left border-right"
            // eslint-disable-next-line react/forbid-dom-props
            style={{ width: 'fit-content', maxWidth: '100%' }}
        >
            <ListTag className={classNames('list-group list-group-horizontal')}>{children}</ListTag>
        </div>
    </div>
)

export const ScrollListVertical: React.FunctionComponent<Props> = ({
    title,
    listTag: ListTag = 'ul',
    className,
    children,
}) => (
    <div className={classNames(className)}>
        <h4 className="font-weight-bold">{title}</h4>
        <div className="overflow-auto">
            <ListTag className={classNames('list-group')}>{children}</ListTag>
        </div>
    </div>
)
