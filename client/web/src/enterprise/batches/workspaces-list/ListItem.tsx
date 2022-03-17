import React from 'react'

import classNames from 'classnames'

import styles from './ListItem.module.scss'

interface ListItemProps {
    /** An optional class to apply to the list item, i.e. if it's selected, stale, etc. */
    className?: string
}

export const ListItem: React.FunctionComponent<ListItemProps> = ({ className, children }) => (
    <li className={classNames('list-group-item', styles.listGroupItem, className)}>{children}</li>
)
