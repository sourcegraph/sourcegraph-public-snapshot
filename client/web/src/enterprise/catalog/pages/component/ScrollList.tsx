import classNames from 'classnames'
import { LocationDescriptor } from 'history'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Badge } from '@sourcegraph/wildcard'

interface Props {
    title: string
    titleLink?: LocationDescriptor
    titleCount?: number
    description?: string | React.ReactFragment
    listTag: 'ol' | 'ul'
    className?: string
    listClassName?: string
}

export const ScrollListHorizontal: React.FunctionComponent<Props> = ({
    title,
    titleLink,
    titleCount,
    description,
    listTag: ListTag = 'ul',
    className,
    listClassName,
    children,
}) => (
    <div className={classNames(className)}>
        <h4 className="font-weight-bold">
            <LinkOrSpan to={titleLink} className="text-body">
                {title}
                {titleCount !== undefined && (
                    <Badge variant="secondary" small={true} pill={true} className="ml-1">
                        {titleCount}
                    </Badge>
                )}
            </LinkOrSpan>
        </h4>
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
    titleLink,
    titleCount,
    description,
    listTag: ListTag = 'ul',
    className,
    listClassName,
    children,
}) => (
    <div className={classNames(className)}>
        <h4 className="font-weight-bold">
            <LinkOrSpan to={titleLink} className="text-body">
                {title}
                {titleCount !== undefined && (
                    <Badge variant="secondary" small={true} pill={true} className="ml-1">
                        {titleCount}
                    </Badge>
                )}
            </LinkOrSpan>
        </h4>
        {description}
        <div className="overflow-auto">
            <ListTag className={classNames('list-group', listClassName)}>{children}</ListTag>
        </div>
    </div>
)
