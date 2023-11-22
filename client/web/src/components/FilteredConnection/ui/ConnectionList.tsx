import React, { type AriaAttributes } from 'react'

import classNames from 'classnames'

import styles from './ConnectionList.module.scss'

interface ConnectionListProps extends AriaAttributes {
    /** list HTML element type. Default is <ul>. */
    as?: 'ul' | 'table' | 'div' | 'ol'

    /** CSS class name for the list element (<ul>, <table>, or <div>). */
    className?: string

    compact?: boolean
}

/**
 * Render a list of FilteredConnection nodes.
 * Can be configured to render as different elements to support alternative representations of data such as through the <table> element.
 */
export const ConnectionList: React.FunctionComponent<React.PropsWithChildren<ConnectionListProps>> = ({
    as: ListComponent = 'ul',
    className,
    children,
    compact,
    ...props
}) => (
    <ListComponent
        className={classNames(styles.normal, compact && styles.compact, className)}
        data-testid="filtered-connection-nodes"
        {...props}
    >
        {children}
    </ListComponent>
)
