import React from 'react'

import classNames from 'classnames'

import { Button, Tooltip } from '@sourcegraph/wildcard'

import styles from './ActionItem.module.scss'

interface ActionItemsProps {
    label?: string
    tooltip?: string
    iconURL?: string
    className?: string
    iconClassName?: string

    onClick?: () => void
}

export const ActionItem: React.FC<ActionItemsProps> = React.memo(props => (
    <Tooltip content={props.tooltip}>
        <Button className={classNames('test-action-item', styles.item, props.className)} onClick={props.onClick}>
            {props.iconURL ? (
                <img src={props.iconURL} alt={props.label} className={classNames(styles.icon, props.iconClassName)} />
            ) : (
                props.label
            )}
        </Button>
    </Tooltip>
))

ActionItem.displayName = 'ActionItem'
