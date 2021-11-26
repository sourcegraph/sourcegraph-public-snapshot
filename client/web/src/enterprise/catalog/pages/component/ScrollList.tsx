import classNames from 'classnames'
import React from 'react'

interface Props {
    title: string
    description?: string | React.ReactFragment
    listTag: 'ol' | 'ul'
    className?: string
    listClassName?: string
}

export const ScrollListHorizontal: React.FunctionComponent<Props> = ({
    title,
    description,
    listTag: ListTag = 'ul',
    className,
    listClassName,
    children,
}) => (
    <div className={classNames(className)}>
        <h4 className="font-weight-bold">{title}</h4>
        {description}
        <div
            className="overflow-auto rounded border-left border-right"
            // eslint-disable-next-line react/forbid-dom-props
            style={{ width: 'fit-content', maxWidth: '100%' }}
        >
            <ListTag className={classNames('list-group list-group-horizontal', listClassName)}>{children}</ListTag>
        </div>
    </div>
)

export const ScrollListVertical: React.FunctionComponent<Props> = ({
    title,
    description,
    listTag: ListTag = 'ul',
    className,
    listClassName,
    children,
}) => (
    <div className={classNames(className)}>
        <h4 className="font-weight-bold">{title}</h4>
        {description}
        <div className="overflow-auto">
            <ListTag className={classNames('list-group', listClassName)}>{children}</ListTag>
        </div>
    </div>
)
