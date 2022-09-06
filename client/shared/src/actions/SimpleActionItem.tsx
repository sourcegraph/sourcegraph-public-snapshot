import * as React from 'react'

import classNames from 'classnames'

import { Button, Tooltip } from '@sourcegraph/wildcard'

import styles from './SimpleActionItem.module.scss'

export interface SimpleActionItemProps {
    isActive?: boolean
    iconURL: string
    tooltip: string
    onClick: (event: React.MouseEvent<HTMLElement>) => void
}

export const SimpleActionItem: React.FunctionComponent<SimpleActionItemProps> = props => {
    const { isActive, iconURL, tooltip, onClick, ...otherProps } = props
    return (
        <Tooltip content={tooltip}>
            <Button
                className={classNames(styles.simpleActionItem, isActive && styles.simpleActionItemActive)}
                onClick={onClick}
                aria-label={tooltip}
                {...otherProps}
            >
                <img src={iconURL} alt={tooltip || ''} />
            </Button>
        </Tooltip>
    )
}
