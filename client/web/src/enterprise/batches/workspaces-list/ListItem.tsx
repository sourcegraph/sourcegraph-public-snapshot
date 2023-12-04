import React from 'react'

import classNames from 'classnames'

import { Button } from '@sourcegraph/wildcard'

import styles from './ListItem.module.scss'

interface ListItemProps {
    /** An optional class to apply to the list item, i.e. if it's selected, stale, etc. */
    className?: string
    /** An optional handler for when the list item is clicked. */
    onClick?: () => void
}

export const ListItem: React.FunctionComponent<React.PropsWithChildren<ListItemProps>> = ({
    className,
    children,
    onClick,
}) => {
    if (!onClick) {
        return <li className={classNames(styles.listGroupItem, className)}>{children}</li>
    }
    return (
        <li className={styles.listGroupItem}>
            <Button className={classNames(styles.button, className)} onClick={onClick}>
                {children}
            </Button>
        </li>
    )
}
